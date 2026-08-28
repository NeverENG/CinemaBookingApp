package handler

import (
	"net/http"

	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/resp"
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/service"
	"github.com/gin-gonic/gin"
)

// PointsHandler 用户积分查询。
type PointsHandler struct {
	points *service.PointsSvc
}

func NewPointsHandler(points *service.PointsSvc) *PointsHandler {
	return &PointsHandler{points: points}
}

// Get GET /api/v1/me/points
func (h *PointsHandler) Get(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "invalid user")
		return
	}
	view, err := h.points.GetPoints(c.Request.Context(), userID)
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, view)
}

type exchangeRequest struct {
	TemplateID int64 `json:"template_id" binding:"required"`
}

// Exchange POST /api/v1/me/points/exchange
func (h *PointsHandler) Exchange(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "invalid user")
		return
	}
	var req exchangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	res, err := h.points.Exchange(c.Request.Context(), userID, req.TemplateID)
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, res)
}

// ListRedeemable GET /api/v1/coupons/redeemable
func (h *PointsHandler) ListRedeemable(c *gin.Context) {
	templates, err := h.points.ListRedeemable(c.Request.Context())
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, templates)
}
