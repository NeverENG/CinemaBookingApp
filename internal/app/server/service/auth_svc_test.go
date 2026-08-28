package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/pkg/crypto"
	"github.com/NeverENG/CinemaBookingApp/internal/pkg/jwt"
)

type fakeAdminRepo struct {
	admins map[string]*domain.Admin
}

func (f *fakeAdminRepo) GetByUsername(ctx context.Context, username string) (*domain.Admin, error) {
	if a, ok := f.admins[username]; ok {
		return a, nil
	}
	return nil, domain.ErrAdminNotFound
}

func (f *fakeAdminRepo) Count(ctx context.Context) (int64, error) {
	return int64(len(f.admins)), nil
}

func (f *fakeAdminRepo) Create(ctx context.Context, admin *domain.Admin) error {
	if f.admins == nil {
		f.admins = make(map[string]*domain.Admin)
	}
	admin.ID = int64(len(f.admins) + 1)
	f.admins[admin.Username] = admin
	return nil
}

type fakeRoleRepo struct {
	roles map[string]*domain.Role
}

func (f *fakeRoleRepo) GetByID(ctx context.Context, id int64) (*domain.Role, error) {
	for _, r := range f.roles {
		if r.ID == id {
			return r, nil
		}
	}
	return nil, domain.ErrRoleNotFound
}

func (f *fakeRoleRepo) GetByCode(ctx context.Context, code string) (*domain.Role, error) {
	if r, ok := f.roles[code]; ok {
		return r, nil
	}
	return nil, domain.ErrRoleNotFound
}

func (f *fakeRoleRepo) Ensure(ctx context.Context, roles []domain.Role) error {
	for _, r := range roles {
		if _, ok := f.roles[r.Code]; !ok {
			f.roles[r.Code] = &r
		}
	}
	return nil
}

func newAuthTestSvc(users *fakeUserRepo, admins *fakeAdminRepo, roles *fakeRoleRepo) *AuthSvc {
	return NewAuthSvc(users, admins, roles, jwt.New("test-secret", time.Hour))
}

func TestUserLogin(t *testing.T) {
	hash, _ := crypto.HashPassword("pass123")
	users := &fakeUserRepo{users: map[int64]*domain.User{
		1: {ID: 1, Username: "alice", PasswordHash: hash, Status: "ACTIVE"},
	}}
	svc := newAuthTestSvc(users, &fakeAdminRepo{}, &fakeRoleRepo{})

	token, user, err := svc.UserLogin(context.Background(), "alice", "pass123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if token == "" || user.ID != 1 {
		t.Fatal("unexpected login result")
	}
}

func TestUserLoginWrongPassword(t *testing.T) {
	hash, _ := crypto.HashPassword("pass123")
	users := &fakeUserRepo{users: map[int64]*domain.User{
		1: {ID: 1, Username: "alice", PasswordHash: hash, Status: "ACTIVE"},
	}}
	svc := newAuthTestSvc(users, &fakeAdminRepo{}, &fakeRoleRepo{})

	_, _, err := svc.UserLogin(context.Background(), "alice", "wrong")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAdminLogin(t *testing.T) {
	hash, _ := crypto.HashPassword("admin123")
	admins := &fakeAdminRepo{admins: map[string]*domain.Admin{
		"admin": {ID: 1, Username: "admin", PasswordHash: hash, RoleID: 1, Status: "ACTIVE"},
	}}
	roles := &fakeRoleRepo{roles: map[string]*domain.Role{
		domain.RoleSuperAdmin: {ID: 1, Code: domain.RoleSuperAdmin, Name: "超级管理员"},
	}}
	svc := newAuthTestSvc(&fakeUserRepo{}, admins, roles)

	token, admin, err := svc.AdminLogin(context.Background(), "admin", "admin123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if token == "" || admin.RoleCode != domain.RoleSuperAdmin {
		t.Fatal("unexpected admin login result")
	}
}

func TestEnsureDefaultAdmin(t *testing.T) {
	admins := &fakeAdminRepo{}
	roles := &fakeRoleRepo{roles: map[string]*domain.Role{
		domain.RoleSuperAdmin: {ID: 1, Code: domain.RoleSuperAdmin, Name: "超级管理员"},
	}}
	svc := newAuthTestSvc(&fakeUserRepo{}, admins, roles)

	if err := svc.EnsureDefaultAdmin(context.Background()); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if len(admins.admins) != 1 {
		t.Fatal("expected one default admin")
	}
	admin := admins.admins["admin"]
	if !crypto.CheckPassword(admin.PasswordHash, "admin123") {
		t.Fatal("default admin password mismatch")
	}
}
