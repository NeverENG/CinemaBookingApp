package handler

import (
	"net/http"

	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/resp"
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/service"
	"github.com/gin-gonic/gin"
)

// RefundHandler 用户退款接口。
type RefundHandler struct {
	refunds *service.RefundSvc
}

func NewRefundHandler(refunds *service.RefundSvc) *RefundHandler {
	return &RefundHandler{refunds: refunds}
}

type applyRefundRequest struct {
	Reason string `json:"reason"`
}

// Apply POST /api/v1/orders/:order_no/refund
func (h *RefundHandler) Apply(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "invalid user")
		return
	}
	var req applyRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	refund, err := h.refunds.ApplyRefund(c.Request.Context(), userID, service.ApplyRefundInput{
		OrderNo: c.Param("order_no"),
		Reason:  req.Reason,
	})
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, refund)
}

type mockRefundCallbackRequest struct {
	RefundNo string `json:"refund_no" binding:"required"`
}

// MockCallback POST /api/v1/refunds/mock-callback
func (h *RefundHandler) MockCallback(c *gin.Context) {
	var req mockRefundCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.refunds.HandleMockCallback(c.Request.Context(), req.RefundNo); err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, gin.H{"status": "success"})
}
