package service

import (
	"context"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"github.com/NeverENG/CinemaBookingApp/internal/pkg/crypto"
)

const (
	roleUser             = "USER"
	defaultAdminUsername = "admin"
	defaultAdminPassword = "admin123"
)

// AuthSvc 用户/管理员登录与 JWT 签发。
type AuthSvc struct {
	users  port.UserRepo
	admins port.AdminRepo
	roles  port.RoleRepo
	tokens port.TokenManager
}

func NewAuthSvc(users port.UserRepo, admins port.AdminRepo, roles port.RoleRepo, tokens port.TokenManager) *AuthSvc {
	return &AuthSvc{users: users, admins: admins, roles: roles, tokens: tokens}
}

func (s *AuthSvc) UserLogin(ctx context.Context, username, password string) (string, *domain.User, error) {
	user, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return "", nil, domain.ErrInvalidCredentials
	}
	if user.Status != "ACTIVE" || !crypto.CheckPassword(user.PasswordHash, password) {
		return "", nil, domain.ErrInvalidCredentials
	}
	token, err := s.tokens.Generate(user.ID, roleUser)
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}

func (s *AuthSvc) AdminLogin(ctx context.Context, username, password string) (string, *domain.Admin, error) {
	admin, err := s.admins.GetByUsername(ctx, username)
	if err != nil {
		return "", nil, domain.ErrInvalidCredentials
	}
	role, err := s.roles.GetByID(ctx, admin.RoleID)
	if err != nil {
		return "", nil, err
	}
	if admin.Status != "ACTIVE" || !crypto.CheckPassword(admin.PasswordHash, password) {
		return "", nil, domain.ErrInvalidCredentials
	}
	admin.RoleCode = role.Code
	token, err := s.tokens.Generate(admin.ID, role.Code)
	if err != nil {
		return "", nil, err
	}
	return token, admin, nil
}

// EnsureDefaultAdmin 开发引导：无管理员时创建 admin/admin123。
// TODO(你): 上线前改为初始化脚本 + 强制改密。
func (s *AuthSvc) EnsureDefaultAdmin(ctx context.Context) error {
	n, err := s.admins.Count(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	role, err := s.roles.GetByCode(ctx, domain.RoleSuperAdmin)
	if err != nil {
		return err
	}
	hash, err := crypto.HashPassword(defaultAdminPassword)
	if err != nil {
		return err
	}
	return s.admins.Create(ctx, &domain.Admin{
		Username:     defaultAdminUsername,
		PasswordHash: hash,
		Nickname:     "超级管理员",
		RoleID:       role.ID,
		Status:       "ACTIVE",
	})
}
