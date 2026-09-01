package domain

import "time"

// User 用户（下单校验用）。
type User struct {
	ID                int64
	Username          string
	Email             string
	PasswordHash      string
	Nickname          string
	MembershipLevelID int64
	Status            string
	EmailVerifiedAt   *time.Time
}
