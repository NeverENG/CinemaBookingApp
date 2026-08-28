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
