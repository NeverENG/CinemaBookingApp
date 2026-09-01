package domain

// Admin 管理员。
type Admin struct {
	ID           int64
	Username     string
	PasswordHash string
	Nickname     string
	RoleID       int64
	RoleCode     string
	CinemaID     *int64 // SUPER_ADMIN 为空；CINEMA_ADMIN 绑定影院
	Status       string
}

// AdminScope 由 JWT 注入的当前管理员上下文，服务层用它做数据隔离。
type AdminScope struct {
	AdminID  int64
	Role     string
	CinemaID *int64
}

func (s AdminScope) IsCinemaAdmin() bool {
	return s.Role == RoleCinemaAdmin
}

func (s AdminScope) CanManageCinema(cinemaID int64) bool {
	if cinemaID <= 0 {
		return false
	}
	if s.Role == RoleSuperAdmin {
		return true
	}
	return s.Role == RoleCinemaAdmin && s.CinemaID != nil && *s.CinemaID == cinemaID
}
