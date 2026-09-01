package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/resp"
	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
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
	users  port.UserRepo
	admins port.AdminRepo
	roles  port.RoleRepo
}

func NewAuthMiddleware(tokens *jwt.Manager, users port.UserRepo, admins port.AdminRepo, roles port.RoleRepo) *AuthMiddleware {
	return &AuthMiddleware{tokens: tokens, users: users, admins: admins, roles: roles}
}

// User 校验普通用户 JWT。
func (m *AuthMiddleware) User() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := m.parse(c)
		if !ok {
			return
		}
		if claims.Role != domain.RoleUser {
			abort(c, http.StatusForbidden, "forbidden")
			return
		}
		user, err := m.users.GetUserByID(c.Request.Context(), claims.UserID)
		if err != nil {
			m.abortIdentityError(c, err)
			return
		}
		if user.Status != "ACTIVE" {
			abort(c, http.StatusUnauthorized, "account disabled")
			return
		}
		c.Set(CtxUserID, user.ID)
		c.Set(CtxRole, domain.RoleUser)
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
		if !domain.IsAdminRole(claims.Role) {
			abort(c, http.StatusForbidden, "forbidden")
			return
		}
		admin, err := m.admins.GetByID(c.Request.Context(), claims.UserID)
		if err != nil {
			m.abortIdentityError(c, err)
			return
		}
		role, err := m.roles.GetByID(c.Request.Context(), admin.RoleID)
		if err != nil {
			m.abortIdentityError(c, err)
			return
		}
		if admin.Status != "ACTIVE" || role.Status != "ACTIVE" || !domain.IsAdminRole(role.Code) {
			abort(c, http.StatusUnauthorized, "account disabled")
			return
		}
		if role.Code == domain.RoleCinemaAdmin && (admin.CinemaID == nil || *admin.CinemaID <= 0) {
			abort(c, http.StatusUnauthorized, "cinema scope missing")
			return
		}
		if !containsRole(roles, role.Code) {
			abort(c, http.StatusForbidden, "forbidden")
			return
		}
		c.Set(CtxUserID, admin.ID)
		c.Set(CtxRole, role.Code)
		if admin.CinemaID != nil {
			c.Set(CtxCinemaID, *admin.CinemaID)
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
		abort(c, http.StatusUnauthorized, "invalid token")
		return nil, false
	}
	return claims, true
}

func (m *AuthMiddleware) abortIdentityError(c *gin.Context, err error) {
	if errors.Is(err, domain.ErrUserNotFound) || errors.Is(err, domain.ErrAdminNotFound) || errors.Is(err, domain.ErrRoleNotFound) {
		abort(c, http.StatusUnauthorized, "invalid token")
		return
	}
	abort(c, http.StatusInternalServerError, "authentication unavailable")
}

func abort(c *gin.Context, status int, message string) {
	resp.Fail(c, status, message)
	c.Abort()
}

func containsRole(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}
