package handler

import (
	"net/http"

	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/resp"
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/service"
	"github.com/gin-gonic/gin"
)

// AuthHandler 登录接口。
type AuthHandler struct {
	auth *service.AuthSvc
}

func NewAuthHandler(auth *service.AuthSvc) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginResponse struct {
	Token  string `json:"token"`
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
}

// UserLogin POST /api/v1/auth/login
func (h *AuthHandler) UserLogin(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	token, user, err := h.auth.UserLogin(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, loginResponse{Token: token, UserID: user.ID, Role: "USER"})
}

// AdminLogin POST /api/v1/admin/auth/login
func (h *AuthHandler) AdminLogin(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	token, admin, err := h.auth.AdminLogin(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, loginResponse{Token: token, UserID: admin.ID, Role: admin.RoleCode})
}
