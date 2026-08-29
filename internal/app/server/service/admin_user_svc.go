package service

import (
	"context"
	"errors"
	"strings"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"github.com/NeverENG/CinemaBookingApp/internal/pkg/crypto"
)

// AdminUserSvc 管理员账号管理：仅 SUPER_ADMIN 可创建 CINEMA_ADMIN / FINANCE。
type AdminUserSvc struct {
	admins port.AdminRepo
	roles  port.RoleRepo
	logs   port.OperationLogRepo
}

func NewAdminUserSvc(admins port.AdminRepo, roles port.RoleRepo, logs port.OperationLogRepo) *AdminUserSvc {
	return &AdminUserSvc{admins: admins, roles: roles, logs: logs}
}

type CreateAdminInput struct {
	Username string
	Password string
	Nickname string
	Role     string
	CinemaID *int64
}

func (s *AdminUserSvc) Create(ctx context.Context, scope domain.AdminScope, in CreateAdminInput) (*domain.Admin, error) {
	if scope.Role != domain.RoleSuperAdmin {
		return nil, domain.ErrForbidden
	}
	username := strings.ToLower(strings.TrimSpace(in.Username))
	nickname := strings.TrimSpace(in.Nickname)
	if username == "" || len(in.Password) < 6 || nickname == "" {
		return nil, domain.ErrInvalidInput
	}
	switch in.Role {
	case domain.RoleCinemaAdmin:
		if in.CinemaID == nil || *in.CinemaID <= 0 {
			return nil, domain.ErrInvalidInput
		}
	case domain.RoleFinance:
		// 财务可不绑定影院（全局财务），也可绑定影院做数据隔离。
	default:
		return nil, domain.ErrInvalidInput
	}
	if _, err := s.admins.GetByUsername(ctx, username); err == nil {
		return nil, domain.ErrUsernameTaken
	} else if !errors.Is(err, domain.ErrAdminNotFound) {
		return nil, err
	}
	role, err := s.roles.GetByCode(ctx, in.Role)
	if err != nil {
		return nil, err
	}
	hash, err := crypto.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}
	admin := &domain.Admin{
		Username:     username,
		PasswordHash: hash,
		Nickname:     nickname,
		RoleID:       role.ID,
		RoleCode:     role.Code,
		CinemaID:     in.CinemaID,
		Status:       "ACTIVE",
	}
	if err := s.admins.Create(ctx, admin); err != nil {
		return nil, err
	}
	_ = s.logs.Create(ctx, &domain.OperationLog{
		AdminID:    scope.AdminID,
		Action:     "CREATE_ADMIN",
		TargetType: "admin",
		TargetID:   username,
		Detail: map[string]any{
			"role":      in.Role,
			"cinema_id": in.CinemaID,
		},
	})
	return admin, nil
}
