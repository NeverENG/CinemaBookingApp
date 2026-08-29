package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/pkg/jwt"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestAuthEngine() (*gin.Engine, *jwt.Manager) {
	tokens := jwt.New("test-secret", time.Hour)
	authMw := NewAuthMiddleware(tokens)
	r := gin.New()
	r.GET("/user", authMw.User(), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	r.GET("/admin", authMw.Admin(domain.RoleSuperAdmin), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	return r, tokens
}

func requestToken(r *gin.Engine, path, token string) int {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

func TestUserEndpointRejectsAdminToken(t *testing.T) {
	r, tokens := newTestAuthEngine()
	userToken, _ := tokens.Generate(1, domain.RoleUser, nil)
	adminToken, _ := tokens.Generate(2, domain.RoleSuperAdmin, nil)
	financeToken, _ := tokens.Generate(3, domain.RoleFinance, nil)

	if got := requestToken(r, "/user", userToken); got != http.StatusOK {
		t.Fatalf("user token on /user: expected 200, got %d", got)
	}
	if got := requestToken(r, "/user", adminToken); got != http.StatusForbidden {
		t.Fatalf("admin token on /user: expected 403, got %d", got)
	}
	if got := requestToken(r, "/user", financeToken); got != http.StatusForbidden {
		t.Fatalf("finance token on /user: expected 403, got %d", got)
	}
	if got := requestToken(r, "/user", ""); got != http.StatusUnauthorized {
		t.Fatalf("missing token on /user: expected 401, got %d", got)
	}
}

func TestAdminEndpointRoleList(t *testing.T) {
	r, tokens := newTestAuthEngine()
	adminToken, _ := tokens.Generate(2, domain.RoleSuperAdmin, nil)
	financeToken, _ := tokens.Generate(3, domain.RoleFinance, nil)
	userToken, _ := tokens.Generate(1, domain.RoleUser, nil)

	if got := requestToken(r, "/admin", adminToken); got != http.StatusOK {
		t.Fatalf("super admin on /admin: expected 200, got %d", got)
	}
	if got := requestToken(r, "/admin", financeToken); got != http.StatusForbidden {
		t.Fatalf("finance on /admin: expected 403, got %d", got)
	}
	if got := requestToken(r, "/admin", userToken); got != http.StatusForbidden {
		t.Fatalf("user token on /admin: expected 403, got %d", got)
	}
}
