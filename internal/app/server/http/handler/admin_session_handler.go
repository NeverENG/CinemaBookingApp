package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/resp"
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/service"
	"github.com/gin-gonic/gin"
)

// AdminSessionHandler 场次管理 HTTP 层。
type AdminSessionHandler struct {
	sessions *service.AdminSessionSvc
}

func NewAdminSessionHandler(sessions *service.AdminSessionSvc) *AdminSessionHandler {
	return &AdminSessionHandler{sessions: sessions}
}

type sessionRequest struct {
	CinemaID       int64  `json:"cinema_id" binding:"required"`
	HallID         int64  `json:"hall_id" binding:"required"`
	MovieID        int64  `json:"movie_id" binding:"required"`
	StartTime      string `json:"start_time" binding:"required"`
	EndTime        string `json:"end_time" binding:"required"`
	BasePriceCents int64  `json:"base_price_cents" binding:"required,gt=0"`
	PriceRules     string `json:"price_rules"`
}

func (h *AdminSessionHandler) Create(c *gin.Context) {
	adminID, ok := adminIDFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "missing admin")
		return
	}
	var req sessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	in, err := toSessionInput(req)
	if err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	session, err := h.sessions.Create(c.Request.Context(), adminID, in)
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, session)
}

func (h *AdminSessionHandler) UpdatePrice(c *gin.Context) {
	adminID, ok := adminIDFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "missing admin")
		return
	}
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, http.StatusBadRequest, "invalid session id")
		return
	}
	var req struct {
		BasePriceCents int64  `json:"base_price_cents" binding:"required,gt=0"`
		PriceRules     string `json:"price_rules"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.sessions.UpdatePrice(c.Request.Context(), adminID, sessionID, req.BasePriceCents, req.PriceRules); err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, gin.H{"id": sessionID})
}

func (h *AdminSessionHandler) Cancel(c *gin.Context) {
	adminID, ok := adminIDFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "missing admin")
		return
	}
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, http.StatusBadRequest, "invalid session id")
		return
	}
	if err := h.sessions.Cancel(c.Request.Context(), adminID, sessionID); err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, gin.H{"id": sessionID})
}

func toSessionInput(req sessionRequest) (service.SessionInput, error) {
	start, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return service.SessionInput{}, err
	}
	end, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		return service.SessionInput{}, err
	}
	return service.SessionInput{
		CinemaID:       req.CinemaID,
		HallID:         req.HallID,
		MovieID:        req.MovieID,
		StartTime:      start,
		EndTime:        end,
		BasePriceCents: req.BasePriceCents,
		PriceRulesJSON: req.PriceRules,
	}, nil
}
