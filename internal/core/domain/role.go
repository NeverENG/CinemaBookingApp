package domain

// 角色编码。
const (
	RoleSuperAdmin  = "SUPER_ADMIN"
	RoleCinemaAdmin = "CINEMA_ADMIN"
	RoleFinance     = "FINANCE"
)

// Role 角色：RBAC 最小模型。
type Role struct {
	ID          int64
	Code        string
	Name        string
	Permissions []string
	Status      string
}
