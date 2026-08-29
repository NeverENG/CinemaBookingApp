package handler

import (
	"net/http"

	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/resp"
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/service"
	"github.com/gin-gonic/gin"
)

// OrderHandler 订单 HTTP 层：只做参数绑定、鉴权、错误映射。
type OrderHandler struct {
	orders *service.OrderSvc
	views  *service.OrderQuerySvc
}

func NewOrderHandler(orders *service.OrderSvc, views ...*service.OrderQuerySvc) *OrderHandler {
	var query *service.OrderQuerySvc
	if len(views) > 0 {
		query = views[0]
	}
	return &OrderHandler{orders: orders, views: query}
}

type createOrderRequest struct {
	SessionID int64   `json:"session_id" binding:"required"`
	SeatIDs   []int64 `json:"seat_ids" binding:"required,min=1"`
	CouponNo  string  `json:"coupon_no"`
}

type createOrderResponse struct {
	OrderNo   string   `json:"order_no"`
	ExpireAt  string   `json:"expire_at"`
	PaidCents int64    `json:"paid_cents"`
	SeatNos   []string `json:"seat_nos"`
}

// Create POST /api/v1/orders
func (h *OrderHandler) Create(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "invalid user")
		return
	}

	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	order, err := h.orders.CreateOrder(c.Request.Context(), service.CreateOrderInput{
		UserID:    userID,
		SessionID: req.SessionID,
		SeatIDs:   req.SeatIDs,
		CouponNo:  req.CouponNo,
	})
	if err != nil {
		resp.Error(c, err)
		return
	}

	seatNos := make([]string, 0, len(order.Items))
	for _, item := range order.Items {
		seatNos = append(seatNos, item.SeatNo)
	}
	resp.OK(c, createOrderResponse{
		OrderNo:   order.OrderNo,
		ExpireAt:  order.ExpireAt.Format("2006-01-02 15:04:05"),
		PaidCents: order.PaidCents,
		SeatNos:   seatNos,
	})
}

// Get GET /api/v1/orders/:order_no（支付后轮询订单状态）
func (h *OrderHandler) Get(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "invalid user")
		return
	}
	var order any
	var err error
	if h.views != nil {
		order, err = h.views.Get(c.Request.Context(), userID, c.Param("order_no"))
	} else {
		order, err = h.orders.GetOrder(c.Request.Context(), userID, c.Param("order_no"))
	}
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, order)
}

func (h *OrderHandler) List(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "invalid user")
		return
	}
	if h.views != nil {
		orders, err := h.views.List(c.Request.Context(), userID)
		if err != nil {
			resp.Error(c, err)
			return
		}
		resp.OK(c, orders)
		return
	}
	orders, err := h.orders.ListOrders(c.Request.Context(), userID)
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, orders)
}
