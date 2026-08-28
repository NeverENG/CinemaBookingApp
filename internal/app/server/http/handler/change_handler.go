package handler

import (
	"net/http"

	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/resp"
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/service"
	"github.com/gin-gonic/gin"
)

// ChangeHandler 改签接口。
type ChangeHandler struct {
	change *service.ChangeTicketSvc
}

func NewChangeHandler(change *service.ChangeTicketSvc) *ChangeHandler {
	return &ChangeHandler{change: change}
}

type changeRequest struct {
	NewSessionID int64   `json:"new_session_id" binding:"required"`
	NewSeatIDs   []int64 `json:"new_seat_ids" binding:"required,min=1"`
}

// Change POST /api/v1/orders/:order_no/change
func (h *ChangeHandler) Change(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "invalid user")
		return
	}
	var req changeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := h.change.Change(c.Request.Context(), userID, c.Param("order_no"), req.NewSessionID, req.NewSeatIDs)
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, result)
}
