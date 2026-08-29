package domain

import "time"

// LoginGuard 登录防爆破状态：按 (scope, username) 记录失败次数与锁定截止时间。
type LoginGuard struct {
	Scope       string
	Username    string
	FailedCount int
	LockedUntil *time.Time
}
