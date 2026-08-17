package middleware

import (
	"net/http"
	"strings"

	"it-platform-server/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth JWT 认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Header 获取 Token，否则回退到 HttpOnly Cookie（access_token）
		authHeader := c.GetHeader("Authorization")
		tokenString := ""
		if authHeader != "" {
			// 解析 Bearer Token
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "Token 格式错误"})
				c.Abort()
				return
			}
			tokenString = parts[1]
		} else {
			// 浏览器原生请求（如 file-viewer PDF 预览的 Range 直连）无法携带自定义 Header，
			// 从同源 Cookie 回退读取 access token
			tokenString, _ = c.Cookie("access_token")
			if tokenString == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
				c.Abort()
				return
			}
		}
		secret := config.Cfg.Server.JWTSecret
		if secret == "" {
			secret = "default-secret-key"
		}

		// 验证 Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			// 来自 Cookie 的 token 无效时同步清除，避免失效 cookie 持续残留
			if authHeader == "" {
				c.SetCookie("access_token", "", -1, "/api", "", false, true)
			}
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "Token 无效或已过期"})
			c.Abort()
			return
		}

		// 将用户信息存入上下文，并校验 Token 类型
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// 校验 Token 类型必须为 access 或 domain_account（兼容旧 Token：无 type 字段视为 access）
			tokenType, _ := claims["type"].(string)
			if tokenType != "" && tokenType != "access" && tokenType != "domain_account" {
				c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "Token 类型无效"})
				c.Abort()
				return
			}
			c.Set("username", claims["username"])
			c.Set("display_name", claims["display_name"])
		}

		c.Next()
	}
}
