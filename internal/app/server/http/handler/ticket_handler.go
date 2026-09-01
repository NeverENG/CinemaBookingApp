package handler

import (
	"net/http"

	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/resp"
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/service"
	"github.com/gin-gonic/gin"
)

// TicketHandler 影院票券核销 HTTP 层。
type TicketHandler struct {
	verification *service.TicketVerificationSvc
}

func NewTicketHandler(verification *service.TicketVerificationSvc) *TicketHandler {
	return &TicketHandler{verification: verification}
}

type verifyTicketRequest struct {
	TicketNo string `json:"ticket_no" binding:"required"`
}

// Verify POST /api/v1/admin/tickets/verify
func (h *TicketHandler) Verify(c *gin.Context) {
	scope, ok := adminScopeFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "missing admin")
		return
	}
	var req verifyTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := h.verification.Verify(c.Request.Context(), scope, req.TicketNo)
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, result)
}
