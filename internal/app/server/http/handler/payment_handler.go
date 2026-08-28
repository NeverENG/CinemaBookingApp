package handler

import (
	"net/http"

	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/resp"
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/service"
	"github.com/gin-gonic/gin"
)

// PaymentHandler 支付 HTTP 层。
type PaymentHandler struct {
	payments *service.PaymentSvc
}

func NewPaymentHandler(payments *service.PaymentSvc) *PaymentHandler {
	return &PaymentHandler{payments: payments}
}

type createPaymentRequest struct {
	OrderNo string `json:"order_no" binding:"required"`
}

type createPaymentResponse struct {
	TransactionNo   string `json:"transaction_no"`
	AmountCents     int64  `json:"amount_cents"`
	Channel         string `json:"channel"`
	MockCallbackURL string `json:"mock_callback_url"`
}

// Create POST /api/v1/payments
func (h *PaymentHandler) Create(c *gin.Context) {
	var req createPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	payment, err := h.payments.CreatePayment(c.Request.Context(), service.CreatePaymentInput{OrderNo: req.OrderNo})
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, createPaymentResponse{
		TransactionNo:   payment.TransactionNo,
		AmountCents:     payment.AmountCents,
		Channel:         payment.Channel,
		MockCallbackURL: "/api/v1/payments/mock-callback",
	})
}

type mockCallbackRequest struct {
	EventID       string `json:"event_id" binding:"required"`
	TransactionNo string `json:"transaction_no" binding:"required"`
	AmountCents   int64  `json:"amount_cents" binding:"required"`
	Payload       string `json:"payload"`
}

// MockCallback POST /api/v1/payments/mock-callback
// 模拟支付网关回调；真实场景由网关侧调用，不走用户鉴权。
func (h *PaymentHandler) MockCallback(c *gin.Context) {
	var req mockCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.payments.HandleMockCallback(c.Request.Context(), service.MockCallbackInput{
		EventID:       req.EventID,
		TransactionNo: req.TransactionNo,
		AmountCents:   req.AmountCents,
		Payload:       req.Payload,
	}); err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, gin.H{"status": "success"})
}

type mockPayRequest struct {
	TransactionNo string `json:"transaction_no" binding:"required"`
}

// MockPay POST /api/v1/payments/mock-pay
// 模拟支付页「确认支付」，生成回调事件并走完整出票链路。
func (h *PaymentHandler) MockPay(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "invalid user")
		return
	}
	var req mockPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.payments.MockPay(c.Request.Context(), userID, req.TransactionNo); err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, gin.H{"status": "success"})
}

// GetByOrder GET /api/v1/payments/order/:order_no（支付状态轮询）
func (h *PaymentHandler) GetByOrder(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "invalid user")
		return
	}
	payment, err := h.payments.GetByOrder(c.Request.Context(), userID, c.Param("order_no"))
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, payment)
}
