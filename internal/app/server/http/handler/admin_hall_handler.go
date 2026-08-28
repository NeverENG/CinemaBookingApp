package handler

import (
	"net/http"
	"strconv"

	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/resp"
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/service"
	"github.com/gin-gonic/gin"
)

// AdminHallHandler 影厅管理 HTTP 层。
type AdminHallHandler struct {
	halls *service.AdminHallSvc
}

func NewAdminHallHandler(halls *service.AdminHallSvc) *AdminHallHandler {
	return &AdminHallHandler{halls: halls}
}

type hallRequest struct {
	CinemaID   int64  `json:"cinema_id" binding:"required"`
	Name       string `json:"name" binding:"required"`
	SeatLayout string `json:"seat_layout" binding:"required"`
}

func (h *AdminHallHandler) Create(c *gin.Context) {
	scope, ok := adminScopeFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "missing admin")
		return
	}
	var req hallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	hall, err := h.halls.Create(c.Request.Context(), scope, service.HallInput{
		CinemaID:   req.CinemaID,
		Name:       req.Name,
		SeatLayout: req.SeatLayout,
	})
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, hall)
}

func (h *AdminHallHandler) List(c *gin.Context) {
	cinemaID, err := strconv.ParseInt(c.Query("cinema_id"), 10, 64)
	if err != nil || cinemaID <= 0 {
		resp.Fail(c, http.StatusBadRequest, "cinema_id required")
		return
	}
	scope, ok := adminScopeFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "missing admin")
		return
	}
	halls, err := h.halls.ListByCinema(c.Request.Context(), scope, cinemaID)
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, halls)
}

func (h *AdminHallHandler) Update(c *gin.Context) {
	scope, ok := adminScopeFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "missing admin")
		return
	}
	hallID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, http.StatusBadRequest, "invalid hall id")
		return
	}
	var req hallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	hall, err := h.halls.Update(c.Request.Context(), scope, hallID, service.HallInput{
		CinemaID:   req.CinemaID,
		Name:       req.Name,
		SeatLayout: req.SeatLayout,
	})
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, hall)
}
