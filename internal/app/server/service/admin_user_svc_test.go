package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/pkg/crypto"
)

func adminUserTestRepos() (*fakeAdminRepo, *fakeRoleRepo, *fakeOperationLogRepo) {
	roles := &fakeRoleRepo{roles: map[string]*domain.Role{
		domain.RoleSuperAdmin:  {ID: 1, Code: domain.RoleSuperAdmin, Name: "超级管理员"},
		domain.RoleCinemaAdmin: {ID: 2, Code: domain.RoleCinemaAdmin, Name: "影院管理员"},
		domain.RoleFinance:     {ID: 3, Code: domain.RoleFinance, Name: "财务"},
	}}
	return &fakeAdminRepo{}, roles, &fakeOperationLogRepo{}
}

func TestAdminUserCreateCinemaAdmin(t *testing.T) {
	admins, roles, logs := adminUserTestRepos()
	svc := NewAdminUserSvc(admins, roles, logs)
	cinemaID := int64(10)
	admin, err := svc.Create(context.Background(), superAdminScope, CreateAdminInput{
		Username: "ZhangSan",
		Password: "pass123",
		Nickname: "张三",
		Role:     domain.RoleCinemaAdmin,
		CinemaID: &cinemaID,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if admin.Username != "zhangsan" || admin.RoleCode != domain.RoleCinemaAdmin || admin.CinemaID == nil || *admin.CinemaID != 10 {
		t.Fatalf("unexpected admin: %+v", admin)
	}
	if !crypto.CheckPassword(admin.PasswordHash, "pass123") {
		t.Fatal("password hash mismatch")
	}
	if len(logs.logs) != 1 || logs.logs[0].Action != "CREATE_ADMIN" {
		t.Fatal("expected audit log")
	}
}

func TestAdminUserCreateForbidden(t *testing.T) {
	admins, roles, logs := adminUserTestRepos()
	svc := NewAdminUserSvc(admins, roles, logs)
	scope := domain.AdminScope{AdminID: 2, Role: domain.RoleCinemaAdmin, CinemaID: int64Ptr(10)}
	if _, err := svc.Create(context.Background(), scope, CreateAdminInput{
		Username: "lisi", Password: "pass123", Nickname: "李四", Role: domain.RoleFinance,
	}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestAdminUserCreateRequiresCinemaForCinemaAdmin(t *testing.T) {
	admins, roles, logs := adminUserTestRepos()
	svc := NewAdminUserSvc(admins, roles, logs)
	if _, err := svc.Create(context.Background(), superAdminScope, CreateAdminInput{
		Username: "wangwu", Password: "pass123", Nickname: "王五", Role: domain.RoleCinemaAdmin,
	}); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestAdminUserCreateDuplicate(t *testing.T) {
	admins, roles, logs := adminUserTestRepos()
	admins.admins = map[string]*domain.Admin{
		"zhangsan": {ID: 1, Username: "zhangsan"},
	}
	svc := NewAdminUserSvc(admins, roles, logs)
	if _, err := svc.Create(context.Background(), superAdminScope, CreateAdminInput{
		Username: "zhangsan", Password: "pass123", Nickname: "张三", Role: domain.RoleFinance,
	}); !errors.Is(err, domain.ErrUsernameTaken) {
		t.Fatalf("expected username taken, got %v", err)
	}
}

func TestAdminUserCreateFinanceGlobal(t *testing.T) {
	admins, roles, logs := adminUserTestRepos()
	svc := NewAdminUserSvc(admins, roles, logs)
	admin, err := svc.Create(context.Background(), superAdminScope, CreateAdminInput{
		Username: "finance1", Password: "pass123", Nickname: "财务一号", Role: domain.RoleFinance,
	})
	if err != nil {
		t.Fatalf("create finance: %v", err)
	}
	if admin.RoleCode != domain.RoleFinance || admin.CinemaID != nil {
		t.Fatalf("unexpected finance admin: %+v", admin)
	}
}

func TestAdminUserListIncludesCinemaName(t *testing.T) {
	admins, roles, logs := adminUserTestRepos()
	createdAt := time.Date(2026, time.September, 2, 10, 30, 0, 0, time.FixedZone("CST", 8*60*60))
	cinemaID := int64(10)
	admins.admins = map[string]*domain.Admin{
		"cinema_ops": {
			ID: 3, Username: "cinema_ops", Nickname: "万象城运营", RoleCode: domain.RoleCinemaAdmin,
			CinemaID: &cinemaID, CinemaName: "LTerm 万象影城", Status: "ACTIVE", CreatedAt: createdAt,
		},
	}
	svc := NewAdminUserSvc(admins, roles, logs)

	views, err := svc.List(context.Background(), superAdminScope)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(views) != 1 || views[0].Nickname != "万象城运营" || views[0].CinemaName != "LTerm 万象影城" || views[0].CinemaID == nil || *views[0].CinemaID != 10 {
		t.Fatalf("unexpected views: %+v", views)
	}
}

func TestAdminUserListForbidden(t *testing.T) {
	admins, roles, logs := adminUserTestRepos()
	svc := NewAdminUserSvc(admins, roles, logs)
	scope := domain.AdminScope{AdminID: 2, Role: domain.RoleCinemaAdmin, CinemaID: int64Ptr(10)}
	if _, err := svc.List(context.Background(), scope); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}
