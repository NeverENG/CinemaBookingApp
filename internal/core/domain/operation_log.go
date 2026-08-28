package domain

// OperationLog 管理端操作审计（只增）。
type OperationLog struct {
	AdminID    int64
	Action     string
	TargetType string
	TargetID   string
	Detail     any
	IP         string
}
