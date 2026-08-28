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
