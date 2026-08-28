package domain

import "time"

// PasswordResetCode 密码重置验证码（只存哈希）。
type PasswordResetCode struct {
	ID        int64
	Email     string
	CodeHash  string
	ExpiresAt time.Time
	UsedAt    *time.Time
}
