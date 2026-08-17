package handlers

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
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

	// LDAP 认证（含密码过期检测）
	userDN, displayName, passwordExpired, expiryInfo, err := ldapAuthenticate(req.Username, password)
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

	// 生成双 Token
	accessToken, err := generateAccessToken(req.Username, displayName)
	if err != nil {
		services.LogLogin(req.Username, displayName, "login_failure", c.ClientIP(), c.Request.UserAgent(), "生成令牌失败")
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成令牌失败"})
		return
	}
	refreshToken, err := generateRefreshToken(req.Username, displayName)
	if err != nil {
		services.LogLogin(req.Username, displayName, "login_failure", c.ClientIP(), c.Request.UserAgent(), "生成刷新令牌失败")
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成刷新令牌失败"})
		return
	}

	// 设置 refreshToken 到 HttpOnly Cookie
	setRefreshTokenCookie(c, refreshToken)
	// 同步 access token 到 Cookie，供浏览器原生请求（如 file-viewer PDF Range 直连）鉴权回退使用
	setAccessTokenCookie(c, accessToken)

	// 记录登录成功日志
	services.LogLogin(req.Username, displayName, "login_success", c.ClientIP(), c.Request.UserAgent(), "登录成功")

	// 构建响应数据
	responseData := gin.H{
		"token":        accessToken,
		"username":     req.Username,
		"display_name": displayName,
	}

	// 添加密码过期信息
	if passwordExpired {
		responseData["password_expired"] = true
		responseData["password_expires_at"] = "已过期"
		responseData["days_remaining"] = 0
		responseData["password_never_expires"] = false
	} else if expiryInfo != nil {
		responseData["password_expired"] = false
		responseData["password_expires_at"] = expiryInfo.PasswordExpiresAt
		responseData["days_remaining"] = expiryInfo.DaysRemaining
		responseData["password_never_expires"] = expiryInfo.PasswordNeverExpires
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功",
		"data":    responseData,
	})
}

// ldapAuthenticate LDAP 认证（含密码过期检测）
// 返回值: userDN, displayName, passwordExpired, expiryInfo, error
func ldapAuthenticate(username, password string) (string, string, bool, *services.PasswordExpiryResult, error) {
	cfg := &config.Cfg.LDAP

	// 创建 LDAP 连接
	l, err := ldap.DialURL(cfg.Server, ldap.DialWithTLSConfig(getTLSConfig(cfg)))
	if err != nil {
		return "", "", false, nil, fmt.Errorf("连接 LDAP 失败: %v", err)
	}
	defer l.Close()

	// 使用服务账号绑定
	err = l.Bind(cfg.Username, cfg.Password)
	if err != nil {
		return "", "", false, nil, fmt.Errorf("LDAP 绑定失败: %v", err)
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
		return "", "", false, nil, fmt.Errorf("搜索用户失败: %v", err)
	}

	if len(sr.Entries) == 0 {
		return "", "", false, nil, fmt.Errorf("用户不存在")
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
		// 检测 AD 密码过期错误码 data 773
		if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
			errStr := err.Error()
			if strings.Contains(errStr, "773") || strings.Contains(errStr, "data 773") {
				// 密码已过期，但仍允许登录
				log.Printf("用户 %s 密码已过期", username)
				return userDN, displayName, true, nil, nil
			}
			// 账号被锁定 data 775
			if strings.Contains(errStr, "775") || strings.Contains(errStr, "data 775") {
				return "", "", false, nil, fmt.Errorf("账号已被锁定，请联系管理员")
			}
		}
		return "", "", false, nil, fmt.Errorf("密码错误")
	}

	// 认证成功，查询密码过期信息
	expiryInfo, err := services.GetPasswordExpiry(username)
	if err != nil {
		log.Printf("查询用户 %s 密码过期信息失败: %v", username, err)
		// 不影响登录，返回空过期信息
		return userDN, displayName, false, nil, nil
	}

	return userDN, displayName, false, expiryInfo, nil
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

// getJWTSecret 获取 JWT 密钥
func getJWTSecret() string {
	secret := config.Cfg.Server.JWTSecret
	if secret == "" {
		secret = "default-secret-key"
	}
	return secret
}

// generateAccessToken 生成访问 Token（短期）
func generateAccessToken(username, displayName string) (string, error) {
	secret := getJWTSecret()
	expiry := config.Cfg.Server.AccessTokenExpiry

	claims := jwt.MapClaims{
		"type":         "access",
		"username":     username,
		"display_name": displayName,
		"exp":          time.Now().Add(time.Duration(expiry) * time.Minute).Unix(),
		"iat":          time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// generateRefreshToken 生成刷新 Token（长期）
func generateRefreshToken(username, displayName string) (string, error) {
	secret := getJWTSecret()
	expiry := config.Cfg.Server.RefreshTokenExpiry

	claims := jwt.MapClaims{
		"type":         "refresh",
		"username":     username,
		"display_name": displayName,
		"exp":          time.Now().Add(time.Duration(expiry) * 24 * time.Hour).Unix(),
		"iat":          time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// setRefreshTokenCookie 设置 refreshToken 到 HttpOnly Cookie
func setRefreshTokenCookie(c *gin.Context, refreshToken string) {
	expiry := config.Cfg.Server.RefreshTokenExpiry
	c.SetCookie("refresh_token", refreshToken, expiry*86400, "/api", "", false, true)
}

// setAccessTokenCookie 设置 accessToken 到 HttpOnly Cookie
// 路径用 "/" 覆盖所有受保护接口；Secure 与 refresh_token 保持一致（开发环境为 http）
func setAccessTokenCookie(c *gin.Context, accessToken string) {
	expiry := config.Cfg.Server.AccessTokenExpiry
	c.SetCookie("access_token", accessToken, expiry*60, "/", "", false, true)
}

// ClearAccessTokenCookie 清除 accessToken Cookie
func ClearAccessTokenCookie(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", false, true)
}

// ClearRefreshTokenCookie 清除 refreshToken Cookie（导出供其他包使用）
func ClearRefreshTokenCookie(c *gin.Context) {
	c.SetCookie("refresh_token", "", -1, "/api", "", false, true)
}

// RefreshToken 刷新 accessToken
func RefreshToken(c *gin.Context) {
	// 从 Cookie 中获取 refreshToken
	refreshTokenStr, err := c.Cookie("refresh_token")
	if err != nil || refreshTokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未找到刷新令牌"})
		return
	}

	secret := getJWTSecret()

	// 解析 refreshToken
	token, err := jwt.Parse(refreshTokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		ClearRefreshTokenCookie(c)
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "刷新令牌无效或已过期"})
		return
	}

	// 校验 Token 类型必须为 refresh
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		ClearRefreshTokenCookie(c)
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "刷新令牌类型无效"})
		return
	}
	tokenType, _ := claims["type"].(string)
	if tokenType != "refresh" {
		ClearRefreshTokenCookie(c)
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "刷新令牌类型无效"})
		return
	}

	username, _ := claims["username"].(string)
	displayName, _ := claims["display_name"].(string)

	// 签发新的 accessToken
	newAccessToken, err := generateAccessToken(username, displayName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成令牌失败"})
		return
	}

	// 同步刷新 access_token Cookie，保持浏览器原生请求鉴权可用
	setAccessTokenCookie(c, newAccessToken)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "刷新成功",
		"data": gin.H{
			"token": newAccessToken,
		},
	})
}
