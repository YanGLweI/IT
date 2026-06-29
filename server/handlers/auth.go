package handlers

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"it-platform-server/config"
	"it-platform-server/services"

	"github.com/gin-gonic/gin"
	"github.com/go-ldap/ldap/v3"
	"github.com/golang-jwt/jwt/v5"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login LDAP 登录
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户名和密码不能为空"})
		return
	}

	// RSA解密密码
	password, err := DecryptPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "密码解密失败"})
		return
	}

	// LDAP 认证
	userDN, displayName, err := ldapAuthenticate(req.Username, password)
	if err != nil {
		// 记录登录失败日志
		services.LogLogin(req.Username, "", "login_failure", c.ClientIP(), c.Request.UserAgent(), "认证失败: "+err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "认证失败: " + err.Error()})
		return
	}

	// 检查安全组
	if err := checkSecurityGroup(userDN); err != nil {
		// 记录登录失败日志
		services.LogLogin(req.Username, displayName, "login_failure", c.ClientIP(), c.Request.UserAgent(), "无权限访问: "+err.Error())
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权限访问: " + err.Error()})
		return
	}

	// 生成 JWT Token
	token, err := generateJWT(req.Username, displayName)
	if err != nil {
		// 记录登录失败日志
		services.LogLogin(req.Username, displayName, "login_failure", c.ClientIP(), c.Request.UserAgent(), "生成令牌失败")
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成令牌失败"})
		return
	}

	// 记录登录成功日志
	services.LogLogin(req.Username, displayName, "login_success", c.ClientIP(), c.Request.UserAgent(), "登录成功")

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "登录成功",
		"data": gin.H{
			"token":        token,
			"username":     req.Username,
			"display_name": displayName,
		},
	})
}

// ldapAuthenticate LDAP 认证
func ldapAuthenticate(username, password string) (string, string, error) {
	cfg := &config.Cfg.LDAP

	// 创建 LDAP 连接
	l, err := ldap.DialURL(cfg.Server, ldap.DialWithTLSConfig(getTLSConfig(cfg)))
	if err != nil {
		return "", "", fmt.Errorf("连接 LDAP 失败: %v", err)
	}
	defer l.Close()

	// 使用服务账号绑定
	err = l.Bind(cfg.Username, cfg.Password)
	if err != nil {
		return "", "", fmt.Errorf("LDAP 绑定失败: %v", err)
	}

	// 搜索用户
	searchRequest := ldap.NewSearchRequest(
		cfg.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(cfg.UserFilter, username),
		[]string{"dn", "displayName", "cn", "sAMAccountName", "memberOf"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		return "", "", fmt.Errorf("搜索用户失败: %v", err)
	}

	if len(sr.Entries) == 0 {
		return "", "", fmt.Errorf("用户不存在")
	}

	userEntry := sr.Entries[0]
	userDN := userEntry.DN
	displayName := userEntry.GetAttributeValue("displayName")
	if displayName == "" {
		displayName = userEntry.GetAttributeValue("cn")
	}
	if displayName == "" {
		displayName = username
	}

	// 使用用户密码绑定验证密码
	err = l.Bind(userDN, password)
	if err != nil {
		return "", "", fmt.Errorf("密码错误")
	}

	return userDN, displayName, nil
}

// checkSecurityGroup 检查用户是否属于安全组
func checkSecurityGroup(userDN string) error {
	cfg := &config.Cfg.LDAP

	// 创建 LDAP 连接
	l, err := ldap.DialURL(cfg.Server, ldap.DialWithTLSConfig(getTLSConfig(cfg)))
	if err != nil {
		return fmt.Errorf("连接 LDAP 失败: %v", err)
	}
	defer l.Close()

	// 使用服务账号绑定
	err = l.Bind(cfg.Username, cfg.Password)
	if err != nil {
		return fmt.Errorf("LDAP 绑定失败: %v", err)
	}

	// 搜索安全组成员
	searchRequest := ldap.NewSearchRequest(
		cfg.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=group)(member=%s))", ldap.EscapeFilter(userDN)),
		[]string{"dn"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		return fmt.Errorf("查询安全组失败: %v", err)
	}

	// 检查用户是否属于指定的安全组
	targetGroupDN := strings.ToLower(cfg.SecurityGroupDN)
	for _, entry := range sr.Entries {
		if strings.ToLower(entry.DN) == targetGroupDN {
			return nil
		}
	}

	return fmt.Errorf("用户不在允许的安全组中")
}

// getTLSConfig 获取 TLS 配置
func getTLSConfig(cfg *config.LDAPConfig) *tls.Config {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.Insecure,
		ServerName:         "10.60.254.252",
	}

	// 加载 CA 证书
	if cfg.CertPath != "" {
		caCert, err := os.ReadFile(cfg.CertPath)
		if err == nil {
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig.RootCAs = caCertPool
		}
	}

	return tlsConfig
}

// generateJWT 生成 JWT Token
func generateJWT(username, displayName string) (string, error) {
	secret := config.Cfg.Server.JWTSecret
	if secret == "" {
		secret = "default-secret-key"
	}

	claims := jwt.MapClaims{
		"username":     username,
		"display_name": displayName,
		"exp":          time.Now().Add(24 * time.Hour).Unix(),
		"iat":          time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
