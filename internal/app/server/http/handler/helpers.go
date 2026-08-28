package handler

import (
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/middleware"
	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
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

// adminScopeFrom 组装管理员上下文（ID + 角色 + 影院）。
func adminScopeFrom(c *gin.Context) (domain.AdminScope, bool) {
	adminID, ok := adminIDFrom(c)
	if !ok {
		return domain.AdminScope{}, false
	}
	scope := domain.AdminScope{AdminID: adminID}
	if v, exists := c.Get(middleware.CtxRole); exists {
		if role, ok := v.(string); ok {
			scope.Role = role
		}
	}
	if v, exists := c.Get(middleware.CtxCinemaID); exists {
		if cinemaID, ok := v.(int64); ok {
			scope.CinemaID = &cinemaID
		}
	}
	return scope, true
}
