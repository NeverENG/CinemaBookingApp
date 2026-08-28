package domain

// MembershipLevel 会员等级：按累计赠送积分升级，只升不降。
type MembershipLevel struct {
	ID             int64
	LevelCode      string
	Name           string
	MinTotalPoints int
	DiscountBp     int
	Status         string
}

// MembershipLevelLog 等级变更记录。
type MembershipLevelLog struct {
	ID          int64
	UserID      int64
	FromLevelID *int64
	ToLevelID   int64
	ChangeType  string
	Reason      string
}
