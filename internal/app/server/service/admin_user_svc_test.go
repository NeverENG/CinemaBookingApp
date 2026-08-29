package service

import (
	"context"
	"errors"
	"testing"

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
