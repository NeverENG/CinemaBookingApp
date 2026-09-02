package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/resp"
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/service"
	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/gin-gonic/gin"
)

// AdminSessionHandler 场次管理 HTTP 层。
type AdminSessionHandler struct {
	sessions *service.AdminSessionSvc
	views    *service.SeatMapSvc
}

func NewAdminSessionHandler(sessions *service.AdminSessionSvc, views *service.SeatMapSvc) *AdminSessionHandler {
	return &AdminSessionHandler{sessions: sessions, views: views}
}

// List GET /api/v1/admin/sessions?cinema_id=
func (h *AdminSessionHandler) List(c *gin.Context) {
	scope, ok := adminScopeFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "missing admin")
		return
	}
	cinemaID := int64(0)
	if raw := c.Query("cinema_id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			resp.Fail(c, http.StatusBadRequest, "invalid cinema id")
			return
		}
		cinemaID = parsed
	}
	if scope.Role == domain.RoleCinemaAdmin {
		if scope.CinemaID == nil || (cinemaID > 0 && *scope.CinemaID != cinemaID) {
			resp.Fail(c, http.StatusForbidden, "forbidden")
			return
		}
		cinemaID = *scope.CinemaID
	}
	sessions, err := h.views.ListSessions(c.Request.Context(), 0, cinemaID)
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, sessions)
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
	scope, ok := adminScopeFrom(c)
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
	session, err := h.sessions.Create(c.Request.Context(), scope, in)
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, session)
}

func (h *AdminSessionHandler) UpdatePrice(c *gin.Context) {
	scope, ok := adminScopeFrom(c)
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
	if err := h.sessions.UpdatePrice(c.Request.Context(), scope, sessionID, req.BasePriceCents, req.PriceRules); err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, gin.H{"id": sessionID})
}

func (h *AdminSessionHandler) Cancel(c *gin.Context) {
	scope, ok := adminScopeFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "missing admin")
		return
	}
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, http.StatusBadRequest, "invalid session id")
		return
	}
	if err := h.sessions.Cancel(c.Request.Context(), scope, sessionID); err != nil {
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
