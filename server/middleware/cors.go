package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Cors 跨域中间件
func Cors() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080", "http://localhost:8081", "http://10.60.254.127:8080", "http://localhost", "http://it.com", "http://10.60.254.124", "http://10.60.1.191:8081", "https://localhost:8080", "https://localhost:8081", "https://10.60.254.127:8080", "https://10.60.254.124", "https://it.com", "http://localhost:8082"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Dual-Control-Token"},
		ExposeHeaders:    []string{"Content-Length", "Content-Disposition"},
		AllowCredentials: true,
	})
}
