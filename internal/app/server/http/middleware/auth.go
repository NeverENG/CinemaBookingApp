package middleware

import (
	"net/http"
	"strings"

	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/resp"
	"github.com/NeverENG/CinemaBookingApp/internal/pkg/jwt"
	"github.com/gin-gonic/gin"
)

const (
	CtxUserID   = "auth.user_id"
	CtxRole     = "auth.role"
	CtxCinemaID = "auth.cinema_id"
)

// AuthMiddleware JWT 鉴权：用户端与管理员端共用解析，角色由调用方限定。
type AuthMiddleware struct {
	tokens *jwt.Manager
}

func NewAuthMiddleware(tokens *jwt.Manager) *AuthMiddleware {
	return &AuthMiddleware{tokens: tokens}
}

// User 校验普通用户 JWT。
func (m *AuthMiddleware) User() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := m.parse(c)
		if !ok {
			return
		}
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxRole, claims.Role)
		if claims.CinemaID != nil {
			c.Set(CtxCinemaID, *claims.CinemaID)
		}
		c.Next()
	}
}

// Admin 校验管理员 JWT 且角色在允许列表。
func (m *AuthMiddleware) Admin(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := m.parse(c)
		if !ok {
			return
		}
		if !containsRole(roles, claims.Role) {
			resp.Fail(c, http.StatusForbidden, "forbidden")
			c.Abort()
			return
		}
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxRole, claims.Role)
		if claims.CinemaID != nil {
			c.Set(CtxCinemaID, *claims.CinemaID)
		}
		c.Next()
	}
}

func (m *AuthMiddleware) parse(c *gin.Context) (*jwt.Claims, bool) {
	tokenStr, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
	if !ok || tokenStr == "" {
		resp.Fail(c, http.StatusUnauthorized, "missing token")
		c.Abort()
		return nil, false
	}
	claims, err := m.tokens.Parse(tokenStr)
	if err != nil {
		resp.Fail(c, http.StatusUnauthorized, "invalid token")
		c.Abort()
		return nil, false
	}
	return claims, true
}

func containsRole(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}
