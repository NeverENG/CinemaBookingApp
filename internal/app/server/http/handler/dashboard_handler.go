package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/resp"
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/service"
	"github.com/gin-gonic/gin"
)

// DashboardHandler 数据看板（管理端）。
type DashboardHandler struct {
	box *service.BoxOfficeSvc
}

func NewDashboardHandler(box *service.BoxOfficeSvc) *DashboardHandler {
	return &DashboardHandler{box: box}
}

// Trend GET /api/v1/dashboard/box-office
func (h *DashboardHandler) Trend(c *gin.Context) {
	scope, ok := adminScopeFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "missing admin")
		return
	}
	q, err := parseDashboardQuery(c)
	if err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	rows, err := h.box.Trend(c.Request.Context(), scope, q)
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, rows)
}

// ByMovie GET /api/v1/dashboard/box-office/by-movie
func (h *DashboardHandler) ByMovie(c *gin.Context) {
	scope, ok := adminScopeFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "missing admin")
		return
	}
	q, err := parseDashboardQuery(c)
	if err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	rows, err := h.box.ByMovie(c.Request.Context(), scope, q)
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, rows)
}

// ByCinema GET /api/v1/dashboard/box-office/by-cinema
func (h *DashboardHandler) ByCinema(c *gin.Context) {
	scope, ok := adminScopeFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "missing admin")
		return
	}
	q, err := parseDashboardQuery(c)
	if err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	rows, err := h.box.ByCinema(c.Request.Context(), scope, q)
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, rows)
}

// Reconcile POST /api/v1/dashboard/box-office/reconcile（由 ledger 重建聚合）
func (h *DashboardHandler) Reconcile(c *gin.Context) {
	if err := h.box.Reconcile(c.Request.Context()); err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, gin.H{"status": "ok"})
}

func parseDashboardQuery(c *gin.Context) (service.DashboardQuery, error) {
	now := time.Now()
	end := now
	start := now.AddDate(0, 0, -7)
	var err error
	if s := c.Query("start_date"); s != "" {
		start, err = time.Parse("2006-01-02", s)
		if err != nil {
			return service.DashboardQuery{}, err
		}
	}
	if e := c.Query("end_date"); e != "" {
		end, err = time.Parse("2006-01-02", e)
		if err != nil {
			return service.DashboardQuery{}, err
		}
	}
	end = end.Add(24 * time.Hour) // end_date 含当天
	cinemaID, _ := strconv.ParseInt(c.Query("cinema_id"), 10, 64)
	movieID, _ := strconv.ParseInt(c.Query("movie_id"), 10, 64)
	granularity := c.Query("granularity")
	if granularity == "" {
		granularity = "day"
	}
	return service.DashboardQuery{
		StartDate:   start,
		EndDate:     end,
		CinemaID:    cinemaID,
		MovieID:     movieID,
		Granularity: granularity,
	}, nil
}
