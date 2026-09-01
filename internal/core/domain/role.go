package domain

// 角色编码。
const (
	RoleUser        = "USER"
	RoleSuperAdmin  = "SUPER_ADMIN"
	RoleCinemaAdmin = "CINEMA_ADMIN"
	RoleFinance     = "FINANCE"
)

func IsAdminRole(role string) bool {
	switch role {
	case RoleSuperAdmin, RoleCinemaAdmin, RoleFinance:
		return true
	default:
		return false
	}
}

// Role 角色：RBAC 最小模型。
type Role struct {
	ID          int64
	Code        string
	Name        string
	Permissions []string
	Status      string
}
