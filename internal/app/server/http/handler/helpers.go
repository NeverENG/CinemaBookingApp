package handler

import (
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/middleware"
	"github.com/gin-gonic/gin"
)

// adminIDFrom 从 JWT 中间件注入的上下文取管理员 ID。
func adminIDFrom(c *gin.Context) (int64, bool) {
	v, exists := c.Get(middleware.CtxUserID)
	if !exists {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok && id > 0
}
