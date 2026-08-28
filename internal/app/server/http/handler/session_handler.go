package handler

import (
	"net/http"
	"strconv"

	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/resp"
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/service"
	"github.com/gin-gonic/gin"
)

// SessionHandler 用户侧场次/座位图查询（公开只读接口）。
type SessionHandler struct {
	seats *service.SeatMapSvc
}

func NewSessionHandler(seats *service.SeatMapSvc) *SessionHandler {
	return &SessionHandler{seats: seats}
}

// List GET /api/v1/sessions?movie_id=&cinema_id=
func (h *SessionHandler) List(c *gin.Context) {
	movieID, _ := strconv.ParseInt(c.Query("movie_id"), 10, 64)
	cinemaID, _ := strconv.ParseInt(c.Query("cinema_id"), 10, 64)
	sessions, err := h.seats.ListSessions(c.Request.Context(), movieID, cinemaID)
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, sessions)
}

// GetSeatMap GET /api/v1/sessions/:id/seats
func (h *SessionHandler) GetSeatMap(c *gin.Context) {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || sessionID <= 0 {
		resp.Fail(c, http.StatusBadRequest, "invalid session id")
		return
	}
	view, err := h.seats.GetSeatMap(c.Request.Context(), sessionID)
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, view)
}
