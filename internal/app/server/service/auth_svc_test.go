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

func (f *fakeAdminRepo) GetByID(ctx context.Context, id int64) (*domain.Admin, error) {
	for _, admin := range f.admins {
		if admin.ID == id {
			return admin, nil
		}
	}
	return nil, domain.ErrAdminNotFound
}

func (f *fakeAdminRepo) GetByUsername(ctx context.Context, username string) (*domain.Admin, error) {
	if a, ok := f.admins[username]; ok {
		return a, nil
	}
	return nil, domain.ErrAdminNotFound
}

func (f *fakeAdminRepo) List(ctx context.Context) ([]domain.Admin, error) {
	admins := make([]domain.Admin, 0, len(f.admins))
	for _, admin := range f.admins {
		admins = append(admins, *admin)
	}
	return admins, nil
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

func (f *fakeAdminRepo) UpdatePassword(ctx context.Context, adminID int64, passwordHash string) error {
	admin, err := f.GetByID(ctx, adminID)
	if err != nil {
		return err
	}
	admin.PasswordHash = passwordHash
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
	return NewAuthSvc(fakeTxManager{}, users, admins, roles, jwt.New("test-secret", time.Hour), &fakeLoginGuardRepo{}, Bootstrap{
		AdminUsername: "admin",
		AdminPassword: "admin123",
		DemoUsername:  "demo@lterm.test",
		DemoPassword:  "demo123",
	}, &fakeEmailVerificationRepo{}, &fakePasswordResetRepo{}, &fakeMembershipRepo{}, &fakeMailSender{}, false)
}

type fakeLoginGuardRepo struct {
	guards map[string]*domain.LoginGuard
}

func (f *fakeLoginGuardRepo) key(scope, username string) string {
	return scope + ":" + username
}

func (f *fakeLoginGuardRepo) Get(ctx context.Context, scope, username string) (*domain.LoginGuard, error) {
	if f.guards == nil {
		return nil, nil
	}
	return f.guards[f.key(scope, username)], nil
}

func (f *fakeLoginGuardRepo) RecordFailure(ctx context.Context, scope, username string) (int, error) {
	if f.guards == nil {
		f.guards = make(map[string]*domain.LoginGuard)
	}
	key := f.key(scope, username)
	g := f.guards[key]
	if g == nil {
		g = &domain.LoginGuard{Scope: scope, Username: username}
		f.guards[key] = g
	}
	g.FailedCount++
	return g.FailedCount, nil
}

func (f *fakeLoginGuardRepo) Lock(ctx context.Context, scope, username string, until time.Time) error {
	if f.guards == nil {
		f.guards = make(map[string]*domain.LoginGuard)
	}
	key := f.key(scope, username)
	g := f.guards[key]
	if g == nil {
		g = &domain.LoginGuard{Scope: scope, Username: username}
		f.guards[key] = g
	}
	g.FailedCount = maxLoginFailures
	g.LockedUntil = &until
	return nil
}

func (f *fakeLoginGuardRepo) Reset(ctx context.Context, scope, username string) error {
	if f.guards != nil {
		if g := f.guards[f.key(scope, username)]; g != nil {
			g.FailedCount = 0
			g.LockedUntil = nil
		}
	}
	return nil
}

type fakePasswordResetRepo struct {
	codes map[string][]*domain.PasswordResetCode
}

type fakeEmailVerificationRepo struct {
	codes map[string][]*domain.EmailVerificationCode
}

func (f *fakeEmailVerificationRepo) Create(ctx context.Context, code *domain.EmailVerificationCode) error {
	if f.codes == nil {
		f.codes = make(map[string][]*domain.EmailVerificationCode)
	}
	code.ID = int64(len(f.codes[code.Email]) + 1)
	f.codes[code.Email] = append(f.codes[code.Email], code)
	return nil
}

func (f *fakeEmailVerificationRepo) FindUnusedByEmail(ctx context.Context, email string) (*domain.EmailVerificationCode, error) {
	codes := f.codes[email]
	for index := len(codes) - 1; index >= 0; index-- {
		code := codes[index]
		if code.UsedAt == nil && code.ExpiresAt.After(time.Now()) {
			return code, nil
		}
	}
	return nil, domain.ErrVerificationCodeInvalid
}

func (f *fakeEmailVerificationRepo) MarkUsed(ctx context.Context, id int64) error {
	for _, codes := range f.codes {
		for _, code := range codes {
			if code.ID == id && code.UsedAt == nil {
				now := time.Now()
				code.UsedAt = &now
				return nil
			}
		}
	}
	return domain.ErrVerificationCodeInvalid
}

func (f *fakePasswordResetRepo) Create(ctx context.Context, code *domain.PasswordResetCode) error {
	if f.codes == nil {
		f.codes = make(map[string][]*domain.PasswordResetCode)
	}
	code.ID = int64(len(f.codes[code.Email]) + 1)
	f.codes[code.Email] = append(f.codes[code.Email], code)
	return nil
}

func (f *fakePasswordResetRepo) FindUnusedByEmail(ctx context.Context, email string) (*domain.PasswordResetCode, error) {
	codes := f.codes[email]
	for i := len(codes) - 1; i >= 0; i-- {
		c := codes[i]
		if c.UsedAt == nil && c.ExpiresAt.After(time.Now()) {
			return c, nil
		}
	}
	return nil, domain.ErrResetCodeInvalid
}

func (f *fakePasswordResetRepo) MarkUsed(ctx context.Context, id int64) error {
	for _, codes := range f.codes {
		for _, c := range codes {
			if c.ID == id {
				now := time.Now()
				c.UsedAt = &now
				return nil
			}
		}
	}
	return domain.ErrResetCodeInvalid
}

type fakeMembershipRepo struct {
	upgrades int
}

func (f *fakeMembershipRepo) UpgradeIfNeeded(ctx context.Context, userID int64) (bool, error) {
	f.upgrades++
	return false, nil
}

type fakeMailSender struct {
	sent []string
}

func (f *fakeMailSender) Send(to, subject, body string) error {
	f.sent = append(f.sent, to)
	return nil
}

func TestUserLogin(t *testing.T) {
	hash, _ := crypto.HashPassword("pass123")
	verifiedAt := time.Now()
	users := &fakeUserRepo{users: map[int64]*domain.User{
		1: {ID: 1, Username: "alice@example.com", Email: "alice@example.com", PasswordHash: hash, Status: "ACTIVE", EmailVerifiedAt: &verifiedAt},
	}}
	svc := newAuthTestSvc(users, &fakeAdminRepo{}, &fakeRoleRepo{})

	token, user, err := svc.UserLogin(context.Background(), "alice@example.com", "pass123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if token == "" || user.ID != 1 {
		t.Fatal("unexpected login result")
	}
}

func TestUserLoginRejectsUnverifiedEmail(t *testing.T) {
	hash, _ := crypto.HashPassword("pass123")
	users := &fakeUserRepo{users: map[int64]*domain.User{
		1: {ID: 1, Username: "alice@example.com", Email: "alice@example.com", PasswordHash: hash, Status: "ACTIVE"},
	}}
	svc := newAuthTestSvc(users, &fakeAdminRepo{}, &fakeRoleRepo{})

	_, _, err := svc.UserLogin(context.Background(), "alice@example.com", "pass123")
	if !errors.Is(err, domain.ErrEmailNotVerified) {
		t.Fatalf("expected ErrEmailNotVerified, got %v", err)
	}
}

func TestUserLoginWrongPassword(t *testing.T) {
	hash, _ := crypto.HashPassword("pass123")
	users := &fakeUserRepo{users: map[int64]*domain.User{
		1: {ID: 1, Username: "alice@example.com", Email: "alice@example.com", PasswordHash: hash, Status: "ACTIVE"},
	}}
	svc := newAuthTestSvc(users, &fakeAdminRepo{}, &fakeRoleRepo{})

	_, _, err := svc.UserLogin(context.Background(), "alice@example.com", "wrong")
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
		domain.RoleSuperAdmin: {ID: 1, Code: domain.RoleSuperAdmin, Name: "超级管理员", Status: "ACTIVE"},
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
		domain.RoleSuperAdmin: {ID: 1, Code: domain.RoleSuperAdmin, Name: "超级管理员", Status: "ACTIVE"},
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

func TestEnsureDefaultCinemaAdmin(t *testing.T) {
	admins := &fakeAdminRepo{}
	roles := &fakeRoleRepo{roles: map[string]*domain.Role{
		domain.RoleCinemaAdmin: {ID: 2, Code: domain.RoleCinemaAdmin, Name: "影院管理员", Status: "ACTIVE"},
	}}
	svc := newAuthTestSvc(&fakeUserRepo{}, admins, roles)
	svc.bootstrap.CinemaAdminUsername = "cinema_admin"
	svc.bootstrap.CinemaAdminPassword = "cinema123"
	svc.bootstrap.CinemaAdminCinemaID = 1

	if err := svc.EnsureDefaultCinemaAdmin(context.Background()); err != nil {
		t.Fatalf("ensure cinema admin: %v", err)
	}
	admin := admins.admins["cinema_admin"]
	if admin == nil || admin.CinemaID == nil || *admin.CinemaID != 1 {
		t.Fatalf("unexpected cinema admin: %+v", admin)
	}
	if !crypto.CheckPassword(admin.PasswordHash, "cinema123") {
		t.Fatal("cinema admin password mismatch")
	}
}

func TestRegister(t *testing.T) {
	users := &fakeUserRepo{}
	verifications := &fakeEmailVerificationRepo{}
	svc := newAuthTestSvc(users, &fakeAdminRepo{}, &fakeRoleRepo{})
	svc.verifications = verifications

	code, err := svc.RequestRegistrationCode(context.Background(), "newbie@example.com")
	if err != nil || len(code) != 6 {
		t.Fatalf("request registration code: code=%q err=%v", code, err)
	}
	token, user, err := svc.Register(context.Background(), "newbie@example.com", code, "pass123", "新用户")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if token == "" || user.ID == 0 {
		t.Fatal("expected token and user id")
	}
	if users.users[user.ID].Username != "newbie@example.com" || users.users[user.ID].Email != "newbie@example.com" {
		t.Fatal("user not stored")
	}
	if users.users[user.ID].EmailVerifiedAt == nil {
		t.Fatal("expected verified email timestamp")
	}
	verification, _ := verifications.FindUnusedByEmail(context.Background(), "newbie@example.com")
	if verification != nil {
		t.Fatal("verification code should be consumed")
	}
}

func TestRegisterDuplicate(t *testing.T) {
	hash, _ := crypto.HashPassword("pass123")
	users := &fakeUserRepo{users: map[int64]*domain.User{
		1: {ID: 1, Username: "alice@example.com", Email: "alice@example.com", PasswordHash: hash, Status: "ACTIVE"},
	}}
	svc := newAuthTestSvc(users, &fakeAdminRepo{}, &fakeRoleRepo{})

	_, _, err := svc.Register(context.Background(), "alice@example.com", "000000", "pass123", "A")
	if !errors.Is(err, domain.ErrUsernameTaken) {
		t.Fatalf("expected ErrUsernameTaken, got %v", err)
	}
}

func TestRegisterWeakPassword(t *testing.T) {
	svc := newAuthTestSvc(&fakeUserRepo{}, &fakeAdminRepo{}, &fakeRoleRepo{})

	_, _, err := svc.Register(context.Background(), "newbie@example.com", "000000", "123", "新用户")
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestRegisterRejectsInvalidVerificationCode(t *testing.T) {
	users := &fakeUserRepo{}
	svc := newAuthTestSvc(users, &fakeAdminRepo{}, &fakeRoleRepo{})

	if _, err := svc.RequestRegistrationCode(context.Background(), "newbie@example.com"); err != nil {
		t.Fatalf("request code: %v", err)
	}
	_, _, err := svc.Register(context.Background(), "newbie@example.com", "000000", "pass123", "新用户")
	if !errors.Is(err, domain.ErrVerificationCodeInvalid) {
		t.Fatalf("expected ErrVerificationCodeInvalid, got %v", err)
	}
}

func TestChangePassword(t *testing.T) {
	hash, _ := crypto.HashPassword("old123")
	users := &fakeUserRepo{users: map[int64]*domain.User{
		1: {ID: 1, Username: "alice@example.com", Email: "alice@example.com", PasswordHash: hash, Status: "ACTIVE"},
	}}
	svc := newAuthTestSvc(users, &fakeAdminRepo{}, &fakeRoleRepo{})

	if err := svc.ChangePassword(context.Background(), 1, "old123", "new123"); err != nil {
		t.Fatalf("change password: %v", err)
	}
	if !crypto.CheckPassword(users.users[1].PasswordHash, "new123") {
		t.Fatal("password not updated")
	}
}

func TestRequestAndResetPassword(t *testing.T) {
	hash, _ := crypto.HashPassword("old123")
	users := &fakeUserRepo{users: map[int64]*domain.User{
		1: {ID: 1, Username: "alice@example.com", Email: "alice@example.com", PasswordHash: hash, Status: "ACTIVE"},
	}}
	resets := &fakePasswordResetRepo{}
	sender := &fakeMailSender{}
	svc := NewAuthSvc(fakeTxManager{}, users, &fakeAdminRepo{}, &fakeRoleRepo{}, jwt.New("test-secret", time.Hour), &fakeLoginGuardRepo{}, Bootstrap{
		AdminUsername: "admin", AdminPassword: "admin123", DemoUsername: "demo@lterm.test", DemoPassword: "demo123",
	}, &fakeEmailVerificationRepo{}, resets, &fakeMembershipRepo{}, sender, true)

	devCode, err := svc.RequestPasswordReset(context.Background(), "alice@example.com")
	if err != nil {
		t.Fatalf("request reset: %v", err)
	}
	if devCode != "" {
		t.Fatal("SMTP enabled should not return dev code")
	}
	if len(sender.sent) != 1 || sender.sent[0] != "alice@example.com" {
		t.Fatalf("expected mail sent, got %v", sender.sent)
	}

	if err := svc.ResetPassword(context.Background(), "alice@example.com", devCode, "new123"); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected invalid input without devCode, got %v", err)
	}
}

func TestResetPasswordDevMode(t *testing.T) {
	hash, _ := crypto.HashPassword("old123")
	users := &fakeUserRepo{users: map[int64]*domain.User{
		1: {ID: 1, Username: "alice@example.com", Email: "alice@example.com", PasswordHash: hash, Status: "ACTIVE"},
	}}
	resets := &fakePasswordResetRepo{}
	svc := NewAuthSvc(fakeTxManager{}, users, &fakeAdminRepo{}, &fakeRoleRepo{}, jwt.New("test-secret", time.Hour), &fakeLoginGuardRepo{}, Bootstrap{
		AdminUsername: "admin", AdminPassword: "admin123", DemoUsername: "demo@lterm.test", DemoPassword: "demo123",
	}, &fakeEmailVerificationRepo{}, resets, &fakeMembershipRepo{}, &fakeMailSender{}, false)

	devCode, err := svc.RequestPasswordReset(context.Background(), "alice@example.com")
	if err != nil {
		t.Fatalf("request reset: %v", err)
	}
	if len(devCode) != 6 {
		t.Fatalf("expected 6-digit dev code, got %q", devCode)
	}
	if err := svc.ResetPassword(context.Background(), "alice@example.com", devCode, "new123"); err != nil {
		t.Fatalf("reset: %v", err)
	}
	if !crypto.CheckPassword(users.users[1].PasswordHash, "new123") {
		t.Fatal("password not reset")
	}
}

func TestUserLoginLocked(t *testing.T) {
	hash, _ := crypto.HashPassword("pass123")
	users := &fakeUserRepo{users: map[int64]*domain.User{
		1: {ID: 1, Username: "alice@example.com", PasswordHash: hash, Status: "ACTIVE"},
	}}
	guards := &fakeLoginGuardRepo{}
	until := time.Now().Add(10 * time.Minute)
	_ = guards.Lock(context.Background(), loginScopeUser, "alice@example.com", until)
	svc := newAuthTestSvc(users, &fakeAdminRepo{}, &fakeRoleRepo{})
	svc.guards = guards

	_, _, err := svc.UserLogin(context.Background(), "alice@example.com", "pass123")
	if !errors.Is(err, domain.ErrAccountLocked) {
		t.Fatalf("expected ErrAccountLocked, got %v", err)
	}
}

func TestUserLoginLockoutAfterFailures(t *testing.T) {
	hash, _ := crypto.HashPassword("pass123")
	users := &fakeUserRepo{users: map[int64]*domain.User{
		1: {ID: 1, Username: "alice@example.com", PasswordHash: hash, Status: "ACTIVE"},
	}}
	guards := &fakeLoginGuardRepo{}
	svc := NewAuthSvc(fakeTxManager{}, users, &fakeAdminRepo{}, &fakeRoleRepo{}, jwt.New("test-secret", time.Hour), guards, Bootstrap{
		AdminUsername: "admin", AdminPassword: "admin123", DemoUsername: "demo@lterm.test", DemoPassword: "demo123",
	}, &fakeEmailVerificationRepo{}, &fakePasswordResetRepo{}, &fakeMembershipRepo{}, &fakeMailSender{}, false)

	for i := 0; i < maxLoginFailures; i++ {
		if _, _, err := svc.UserLogin(context.Background(), "alice@example.com", "wrong"); !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Fatalf("attempt %d: expected ErrInvalidCredentials, got %v", i+1, err)
		}
	}
	g, err := guards.Get(context.Background(), loginScopeUser, "alice@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if g == nil || g.LockedUntil == nil || !time.Now().Before(*g.LockedUntil) {
		t.Fatalf("expected account locked after %d failures", maxLoginFailures)
	}

	if _, _, err := svc.UserLogin(context.Background(), "alice@example.com", "pass123"); !errors.Is(err, domain.ErrAccountLocked) {
		t.Fatalf("expected ErrAccountLocked for correct password, got %v", err)
	}
}

func TestAdminLoginSuccessResetsGuard(t *testing.T) {
	hash, _ := crypto.HashPassword("admin123")
	admins := &fakeAdminRepo{admins: map[string]*domain.Admin{
		"admin": {ID: 1, Username: "admin", PasswordHash: hash, RoleID: 1, Status: "ACTIVE"},
	}}
	roles := &fakeRoleRepo{roles: map[string]*domain.Role{
		domain.RoleSuperAdmin: {ID: 1, Code: domain.RoleSuperAdmin, Name: "超级管理员", Status: "ACTIVE"},
	}}
	guards := &fakeLoginGuardRepo{}
	_, _ = guards.RecordFailure(context.Background(), loginScopeAdmin, "admin")
	svc := NewAuthSvc(fakeTxManager{}, &fakeUserRepo{}, admins, roles, jwt.New("test-secret", time.Hour), guards, Bootstrap{
		AdminUsername: "admin", AdminPassword: "admin123", DemoUsername: "demo@lterm.test", DemoPassword: "demo123",
	}, &fakeEmailVerificationRepo{}, &fakePasswordResetRepo{}, &fakeMembershipRepo{}, &fakeMailSender{}, false)

	if _, _, err := svc.AdminLogin(context.Background(), "admin", "admin123"); err != nil {
		t.Fatalf("admin login: %v", err)
	}
	g, err := guards.Get(context.Background(), loginScopeAdmin, "admin")
	if err != nil {
		t.Fatal(err)
	}
	if g == nil || g.FailedCount != 0 || g.LockedUntil != nil {
		t.Fatalf("expected guard reset after success, got %+v", g)
	}
}

func TestEnsureDemoUserUpdatesConfiguredPassword(t *testing.T) {
	oldHash, _ := crypto.HashPassword("oldpass")
	users := &fakeUserRepo{users: map[int64]*domain.User{
		1: {ID: 1, Username: "showcase@example.com", Email: "showcase@example.com", PasswordHash: oldHash, Status: "ACTIVE"},
	}}
	svc := newAuthTestSvc(users, &fakeAdminRepo{}, &fakeRoleRepo{})
	svc.bootstrap.DemoUsername = " Showcase@Example.com "
	svc.bootstrap.DemoPassword = "newpass123"

	if err := svc.EnsureDemoUser(context.Background()); err != nil {
		t.Fatalf("ensure demo user: %v", err)
	}
	if !crypto.CheckPassword(users.users[1].PasswordHash, "newpass123") {
		t.Fatal("configured demo password was not applied")
	}
}
