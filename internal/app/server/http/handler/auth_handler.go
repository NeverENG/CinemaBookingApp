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

type userLoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type adminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginResponse struct {
	Token    string `json:"token"`
	UserID   int64  `json:"user_id"`
	Role     string `json:"role"`
	CinemaID *int64 `json:"cinema_id,omitempty"`
}

type registerRequest struct {
	Email    string `json:"email" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname" binding:"required"`
}

type verificationCodeResponse struct {
	Status  string `json:"status"`
	DevCode string `json:"dev_code,omitempty"`
}

// Register POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	token, user, err := h.auth.Register(c.Request.Context(), req.Email, req.Code, req.Password, req.Nickname)
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, loginResponse{Token: token, UserID: user.ID, Role: "USER"})
}

// RequestRegistrationCode POST /api/v1/auth/email-verification/request
func (h *AuthHandler) RequestRegistrationCode(c *gin.Context) {
	var req resetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	devCode, err := h.auth.RequestRegistrationCode(c.Request.Context(), req.Email)
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, verificationCodeResponse{Status: "ok", DevCode: devCode})
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ChangePassword POST /api/v1/me/password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "invalid user")
		return
	}
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.auth.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, gin.H{"status": "ok"})
}

type resetRequest struct {
	Email string `json:"email" binding:"required"`
}

// RequestPasswordReset POST /api/v1/auth/password-reset/request
func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	var req resetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	devCode, err := h.auth.RequestPasswordReset(c.Request.Context(), req.Email)
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, verificationCodeResponse{Status: "ok", DevCode: devCode})
}

type resetPasswordRequest struct {
	Email       string `json:"email" binding:"required"`
	Code        string `json:"code" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ResetPassword POST /api/v1/auth/password-reset/reset
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.auth.ResetPassword(c.Request.Context(), req.Email, req.Code, req.NewPassword); err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, gin.H{"status": "ok"})
}

// UserLogin POST /api/v1/auth/login
func (h *AuthHandler) UserLogin(c *gin.Context) {
	var req userLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	token, user, err := h.auth.UserLogin(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, loginResponse{Token: token, UserID: user.ID, Role: "USER"})
}

// AdminLogin POST /api/v1/admin/auth/login
func (h *AuthHandler) AdminLogin(c *gin.Context) {
	var req adminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	token, admin, err := h.auth.AdminLogin(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, loginResponse{Token: token, UserID: admin.ID, Role: admin.RoleCode, CinemaID: admin.CinemaID})
}

// ChangeAdminPassword POST /api/v1/admin/me/password
func (h *AuthHandler) ChangeAdminPassword(c *gin.Context) {
	adminID, ok := adminIDFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "invalid admin")
		return
	}
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.auth.ChangeAdminPassword(c.Request.Context(), adminID, req.OldPassword, req.NewPassword); err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, gin.H{"status": "ok"})
}
