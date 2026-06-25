package handlers

import (
	"net/http"
	"time"

	"it-platform-server/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// VerifyDualControl 双控验证 - 验证另一成员的LDAP账号
func VerifyDualControl(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 不能使用当前登录用户作为双控审批人
	currentUsername, _ := c.Get("username")
	if currentUsername == req.Username {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "不能使用自己的账号进行双控验证"})
		return
	}

	// LDAP验证（仅验证账号密码有效性，不检查安全组）
	_, _, err := ldapAuthenticate(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "验证失败: " + err.Error()})
		return
	}

	// 生成短生命周期的双控token（5分钟有效）
	secret := config.Cfg.Server.JWTSecret
	if secret == "" {
		secret = "default-secret-key"
	}

	claims := jwt.MapClaims{
		"verified_by": req.Username,
		"purpose":     "dual_control",
		"exp":         time.Now().Add(5 * time.Minute).Unix(),
		"iat":         time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成令牌失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"token": tokenString}, "message": "验证成功"})
}
