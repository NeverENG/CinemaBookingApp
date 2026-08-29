package handler

import (
	"net/http"

	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/resp"
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/service"
	"github.com/gin-gonic/gin"
)

// AdminUserHandler 管理员账号管理（仅 SUPER_ADMIN）。
type AdminUserHandler struct {
	users *service.AdminUserSvc
}

func NewAdminUserHandler(users *service.AdminUserSvc) *AdminUserHandler {
	return &AdminUserHandler{users: users}
}

type createAdminRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname" binding:"required"`
	Role     string `json:"role" binding:"required"`
	CinemaID *int64 `json:"cinema_id"`
}

// Create POST /api/v1/admin/admins
func (h *AdminUserHandler) Create(c *gin.Context) {
	scope, ok := adminScopeFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "invalid admin")
		return
	}
	var req createAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	admin, err := h.users.Create(c.Request.Context(), scope, service.CreateAdminInput{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
		Role:     req.Role,
		CinemaID: req.CinemaID,
	})
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, gin.H{
		"id":        admin.ID,
		"username":  admin.Username,
		"role":      admin.RoleCode,
		"cinema_id": admin.CinemaID,
	})
}
