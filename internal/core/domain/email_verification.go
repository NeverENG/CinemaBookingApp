package domain

import "time"

type EmailVerificationCode struct {
	ID        int64
	Email     string
	CodeHash  string
	ExpiresAt time.Time
	UsedAt    *time.Time
}
