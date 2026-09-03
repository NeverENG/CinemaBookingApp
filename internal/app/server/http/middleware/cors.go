package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS 开发用跨域中间件：允许任意来源，生产环境应改为白名单。
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
			c.Header("X-Request-ID", requestID)
		}
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Idempotency-Key, X-Request-ID, X-Callback-Secret")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func CallbackSecret(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if secret == "" || c.GetHeader("X-Callback-Secret") != secret {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.Next()
	}
}
