package middleware

import (
	"context"
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

func int64Ptr(value int64) *int64 {
	return &value
}

type middlewareUserRepo struct {
	users map[int64]*domain.User
}

func (f *middlewareUserRepo) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	if user, ok := f.users[id]; ok {
		return user, nil
	}
	return nil, domain.ErrUserNotFound
}

func (f *middlewareUserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	for _, user := range f.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (f *middlewareUserRepo) Create(ctx context.Context, user *domain.User) error {
	f.users[user.ID] = user
	return nil
}

func (f *middlewareUserRepo) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	user, err := f.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	user.PasswordHash = passwordHash
	return nil
}

type middlewareAdminRepo struct {
	admins map[int64]*domain.Admin
}

func (f *middlewareAdminRepo) GetByID(ctx context.Context, id int64) (*domain.Admin, error) {
	if admin, ok := f.admins[id]; ok {
		return admin, nil
	}
	return nil, domain.ErrAdminNotFound
}

func (f *middlewareAdminRepo) GetByUsername(ctx context.Context, username string) (*domain.Admin, error) {
	for _, admin := range f.admins {
		if admin.Username == username {
			return admin, nil
		}
	}
	return nil, domain.ErrAdminNotFound
}

func (f *middlewareAdminRepo) List(ctx context.Context) ([]domain.Admin, error) {
	admins := make([]domain.Admin, 0, len(f.admins))
	for _, admin := range f.admins {
		admins = append(admins, *admin)
	}
	return admins, nil
}

func (f *middlewareAdminRepo) Count(ctx context.Context) (int64, error) {
	return int64(len(f.admins)), nil
}

func (f *middlewareAdminRepo) Create(ctx context.Context, admin *domain.Admin) error {
	f.admins[admin.ID] = admin
	return nil
}

func (f *middlewareAdminRepo) UpdatePassword(ctx context.Context, adminID int64, passwordHash string) error {
	admin, err := f.GetByID(ctx, adminID)
	if err != nil {
		return err
	}
	admin.PasswordHash = passwordHash
	return nil
}

type middlewareRoleRepo struct {
	roles map[int64]*domain.Role
}

func (f *middlewareRoleRepo) GetByID(ctx context.Context, id int64) (*domain.Role, error) {
	if role, ok := f.roles[id]; ok {
		return role, nil
	}
	return nil, domain.ErrRoleNotFound
}

func (f *middlewareRoleRepo) GetByCode(ctx context.Context, code string) (*domain.Role, error) {
	for _, role := range f.roles {
		if role.Code == code {
			return role, nil
		}
	}
	return nil, domain.ErrRoleNotFound
}

func (f *middlewareRoleRepo) Ensure(ctx context.Context, roles []domain.Role) error {
	for _, role := range roles {
		copy := role
		f.roles[role.ID] = &copy
	}
	return nil
}

func newTestAuthEngine() (*gin.Engine, *jwt.Manager, *middlewareUserRepo, *middlewareAdminRepo, *middlewareRoleRepo) {
	tokens := jwt.New("test-secret", time.Hour)
	users := &middlewareUserRepo{users: map[int64]*domain.User{
		1: {ID: 1, Username: "user", Status: "ACTIVE"},
	}}
	admins := &middlewareAdminRepo{admins: map[int64]*domain.Admin{
		2: {ID: 2, Username: "super", RoleID: 1, Status: "ACTIVE"},
		3: {ID: 3, Username: "finance", RoleID: 3, Status: "ACTIVE"},
		4: {ID: 4, Username: "cinema", RoleID: 2, CinemaID: int64Ptr(10), Status: "ACTIVE"},
	}}
	roles := &middlewareRoleRepo{roles: map[int64]*domain.Role{
		1: {ID: 1, Code: domain.RoleSuperAdmin, Status: "ACTIVE"},
		2: {ID: 2, Code: domain.RoleCinemaAdmin, Status: "ACTIVE"},
		3: {ID: 3, Code: domain.RoleFinance, Status: "ACTIVE"},
	}}
	authMw := NewAuthMiddleware(tokens, users, admins, roles)
	r := gin.New()
	r.GET("/user", authMw.User(), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	r.GET("/admin", authMw.Admin(domain.RoleSuperAdmin), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	return r, tokens, users, admins, roles
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
	r, tokens, _, _, _ := newTestAuthEngine()
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
	r, tokens, _, _, _ := newTestAuthEngine()
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

func TestAdminMiddlewareRechecksCurrentAccount(t *testing.T) {
	r, tokens, _, admins, roles := newTestAuthEngine()
	superToken, _ := tokens.Generate(2, domain.RoleSuperAdmin, nil)

	admins.admins[2].Status = "DISABLED"
	if got := requestToken(r, "/admin", superToken); got != http.StatusUnauthorized {
		t.Fatalf("disabled admin: expected 401, got %d", got)
	}

	admins.admins[2].Status = "ACTIVE"
	roles.roles[1].Status = "DISABLED"
	if got := requestToken(r, "/admin", superToken); got != http.StatusUnauthorized {
		t.Fatalf("disabled role: expected 401, got %d", got)
	}
}
