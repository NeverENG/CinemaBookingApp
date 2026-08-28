package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"github.com/NeverENG/CinemaBookingApp/internal/pkg/crypto"
)

const roleUser = "USER"

// MailSender 邮件发送能力（由 pkg/mailer 实现）。
type MailSender interface {
	Send(to, subject, body string) error
}

// Bootstrap 引导账号配置（由 config 注入，不写死）。
type Bootstrap struct {
	AdminUsername string
	AdminPassword string
	DemoUsername  string
	DemoPassword  string
}

// AuthSvc 用户/管理员登录、注册、密码管理、JWT 签发。
type AuthSvc struct {
	users       port.UserRepo
	admins      port.AdminRepo
	roles       port.RoleRepo
	tokens      port.TokenManager
	bootstrap   Bootstrap
	resets      port.PasswordResetRepo
	members     port.MembershipRepo
	sender      MailSender
	mailEnabled bool
}

func NewAuthSvc(
	users port.UserRepo,
	admins port.AdminRepo,
	roles port.RoleRepo,
	tokens port.TokenManager,
	bootstrap Bootstrap,
	resets port.PasswordResetRepo,
	members port.MembershipRepo,
	sender MailSender,
	mailEnabled bool,
) *AuthSvc {
	return &AuthSvc{
		users:       users,
		admins:      admins,
		roles:       roles,
		tokens:      tokens,
		bootstrap:   bootstrap,
		resets:      resets,
		members:     members,
		sender:      sender,
		mailEnabled: mailEnabled,
	}
}

func (s *AuthSvc) UserLogin(ctx context.Context, username, password string) (string, *domain.User, error) {
	user, err := s.users.GetByUsername(ctx, strings.ToLower(strings.TrimSpace(username)))
	if err != nil {
		return "", nil, domain.ErrInvalidCredentials
	}
	if user.Status != "ACTIVE" || !crypto.CheckPassword(user.PasswordHash, password) {
		return "", nil, domain.ErrInvalidCredentials
	}
	token, err := s.tokens.Generate(user.ID, roleUser, nil)
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}

// Register 用户注册：邮箱即账号 → 建号 → 返回登录态。
func (s *AuthSvc) Register(ctx context.Context, email, password, nickname string) (string, *domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	nickname = strings.TrimSpace(nickname)
	if !validEmail(email) || len(password) < 6 {
		return "", nil, domain.ErrInvalidInput
	}
	if nickname == "" {
		nickname = strings.SplitN(email, "@", 2)[0]
	}
	if _, err := s.users.GetByUsername(ctx, email); err == nil {
		return "", nil, domain.ErrUsernameTaken
	} else if !errors.Is(err, domain.ErrUserNotFound) {
		return "", nil, err
	}
	hash, err := crypto.HashPassword(password)
	if err != nil {
		return "", nil, err
	}
	user := &domain.User{
		Username:     email,
		Email:        email,
		PasswordHash: hash,
		Nickname:     nickname,
		Status:       "ACTIVE",
	}
	if err := s.users.Create(ctx, user); err != nil {
		return "", nil, err
	}
	token, err := s.tokens.Generate(user.ID, roleUser, nil)
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}

// ChangePassword 修改密码：校验旧密码后更新。
func (s *AuthSvc) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if !crypto.CheckPassword(user.PasswordHash, oldPassword) {
		return domain.ErrInvalidCredentials
	}
	if len(newPassword) < 6 {
		return domain.ErrInvalidInput
	}
	hash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.users.UpdatePassword(ctx, userID, hash)
}

// RequestPasswordReset 生成验证码并发邮件；SMTP 未配置时返回 devCode 供本地调试。
func (s *AuthSvc) RequestPasswordReset(ctx context.Context, email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if _, err := s.users.GetByUsername(ctx, email); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", domain.ErrEmailNotRegistered
		}
		return "", err
	}
	code, err := generateResetCode()
	if err != nil {
		return "", err
	}
	hash, err := crypto.HashPassword(code)
	if err != nil {
		return "", err
	}
	if err := s.resets.Create(ctx, &domain.PasswordResetCode{
		Email:     email,
		CodeHash:  hash,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}); err != nil {
		return "", err
	}
	if !s.mailEnabled {
		return code, nil // 开发模式：直接返回验证码
	}
	if err := s.sender.Send(email, "【LTerm】密码重置验证码", "你的验证码是："+code+"，15 分钟内有效。"); err != nil {
		return "", err
	}
	return "", nil
}

// ResetPassword 校验验证码并重置密码。
func (s *AuthSvc) ResetPassword(ctx context.Context, email, code, newPassword string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if len(newPassword) < 6 {
		return domain.ErrInvalidInput
	}
	record, err := s.resets.FindUnusedByEmail(ctx, email)
	if err != nil {
		return err
	}
	if !crypto.CheckPassword(record.CodeHash, code) {
		return domain.ErrResetCodeInvalid
	}
	user, err := s.users.GetByUsername(ctx, email)
	if err != nil {
		return err
	}
	hash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return err
	}
	if err := s.users.UpdatePassword(ctx, user.ID, hash); err != nil {
		return err
	}
	return s.resets.MarkUsed(ctx, record.ID)
}

func validEmail(email string) bool {
	parts := strings.SplitN(email, "@", 2)
	return len(parts) == 2 && parts[0] != "" && strings.Contains(parts[1], ".")
}

func generateResetCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
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
	token, err := s.tokens.Generate(admin.ID, role.Code, admin.CinemaID)
	if err != nil {
		return "", nil, err
	}
	return token, admin, nil
}

// EnsureDefaultAdmin 开发引导：无管理员时创建配置里的默认管理员。
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
	hash, err := crypto.HashPassword(s.bootstrap.AdminPassword)
	if err != nil {
		return err
	}
	return s.admins.Create(ctx, &domain.Admin{
		Username:     s.bootstrap.AdminUsername,
		PasswordHash: hash,
		Nickname:     "超级管理员",
		RoleID:       role.ID,
		Status:       "ACTIVE",
	})
}

// EnsureDemoUser 开发演示用户 demo/demo123（幂等）。
func (s *AuthSvc) EnsureDemoUser(ctx context.Context) error {
	if _, err := s.users.GetByUsername(ctx, s.bootstrap.DemoUsername); err == nil {
		return nil
	} else if !errors.Is(err, domain.ErrUserNotFound) {
		return err
	}
	hash, err := crypto.HashPassword(s.bootstrap.DemoPassword)
	if err != nil {
		return err
	}
	return s.users.Create(ctx, &domain.User{
		Username:     s.bootstrap.DemoUsername,
		Email:        s.bootstrap.DemoUsername,
		PasswordHash: hash,
		Nickname:     "演示用户",
		Status:       "ACTIVE",
	})
}
