package middleware

import (
	"net/http"

	"it-platform-server/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// DualControl 双控验证中间件
// 检查 X-Dual-Control-Token 头中的短生命周期JWT token
func DualControl() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("X-Dual-Control-Token")
		if tokenString == "" {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "需要双控验证"})
			c.Abort()
			return
		}

		secret := config.Cfg.Server.JWTSecret
		if secret == "" {
			secret = "default-secret-key"
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "双控验证无效或已过期"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["purpose"] != "dual_control" {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "双控验证无效"})
			c.Abort()
			return
		}

		c.Set("dual_control_verified_by", claims["verified_by"])
		c.Next()
	}
}
