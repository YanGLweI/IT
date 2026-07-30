package handlers

import (
	"log"
	"net/http"
	"time"

	"it-platform-server/config"
	"it-platform-server/services"
	"it-platform-server/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// DomainAccountLoginRequest 域账号自助平台登录请求
type DomainAccountLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ChangeDomainPasswordRequest 修改域账号密码请求
type ChangeDomainPasswordRequest struct {
	Username    string `json:"username" binding:"required"`
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// DomainAccountLogin 域账号自助平台登录（不做安全组检查）
func DomainAccountLogin(c *gin.Context) {
	var req DomainAccountLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户名和密码不能为空"})
		return
	}

	// 认证并获取密码过期状态
	result, err := services.AuthenticateWithExpiry(req.Username, req.Password)
	if err != nil {
		services.LogLogin(req.Username, "", "login_failure", c.ClientIP(), c.Request.UserAgent(), "域账号认证失败: "+err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	// 密码过期时不签发 token，但返回 expired 状态
	if result.Status == "expired" {
		services.LogLogin(req.Username, "", "login_expired", c.ClientIP(), c.Request.UserAgent(), "域账号密码已过期")
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "密码已过期，请修改密码",
			"data": gin.H{
				"status":   "expired",
				"message":  "密码已过期，请修改密码后重新登录",
				"username": req.Username,
			},
		})
		return
	}

	// 认证成功，生成 JWT（使用域账号自助平台的 token）
	accessToken, err := generateDomainAccountToken(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成令牌失败"})
		return
	}

	services.LogLogin(req.Username, "", "login_success", c.ClientIP(), c.Request.UserAgent(), "域账号自助平台登录成功")

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功",
		"data": gin.H{
			"status":               "ok",
			"token":                accessToken,
			"username":             req.Username,
			"password_expires_at":  result.PasswordExpiryResult.PasswordExpiresAt,
			"days_remaining":       result.PasswordExpiryResult.DaysRemaining,
			"password_never_expires": result.PasswordExpiryResult.PasswordNeverExpires,
		},
	})
}

// GetDomainAccountInfo 获取域账号信息（需要 JWT）
func GetDomainAccountInfo(c *gin.Context) {
	// 从 JWT 上下文获取用户名
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	usernameStr, ok := username.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户信息无效"})
		return
	}

	// 查询密码过期信息
	expiry, err := services.GetPasswordExpiry(usernameStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询账号信息失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"username":             usernameStr,
			"password_expires_at":  expiry.PasswordExpiresAt,
			"days_remaining":       expiry.DaysRemaining,
			"password_never_expires": expiry.PasswordNeverExpires,
		},
	})
}

// ChangeDomainPassword 修改域账号密码
func ChangeDomainPassword(c *gin.Context) {
	var req ChangeDomainPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请填写完整信息"})
		return
	}

	// 1. 校验新旧密码不同
	if err := utils.ValidatePasswordDifferent(req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// 2. 校验新密码复杂度
	if err := utils.ValidatePasswordComplexity(req.NewPassword, req.Username); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// 3. 验证旧密码（如果密码未过期，此步会成功；如果已过期，data 773 也视为通过）
	if err := services.VerifyPassword(req.Username, req.OldPassword); err != nil {
		services.LogLogin(req.Username, "", "password_change_failure", c.ClientIP(), c.Request.UserAgent(), "旧密码验证失败")
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "旧密码验证失败"})
		return
	}

	// 4. 修改密码
	if err := services.ChangePassword(req.Username, req.OldPassword, req.NewPassword); err != nil {
		services.LogLogin(req.Username, "", "password_change_failure", c.ClientIP(), c.Request.UserAgent(), "密码修改失败: "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码修改失败: " + err.Error()})
		return
	}

	// 5. 记录审计日志
	services.LogLogin(req.Username, "", "password_change_success", c.ClientIP(), c.Request.UserAgent(), "域账号密码修改成功")
	log.Printf("域账号 %s 密码修改成功 (IP: %s)", req.Username, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "密码修改成功，请使用新密码重新登录",
	})
}

// generateDomainAccountToken 生成域账号自助平台 JWT
func generateDomainAccountToken(username string) (string, error) {
	secret := config.Cfg.Server.JWTSecret
	if secret == "" {
		secret = "default-secret-key"
	}

	// 自助平台 token 有效期 24 小时
	claims := jwt.MapClaims{
		"type":     "domain_account",
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
