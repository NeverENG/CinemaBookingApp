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

// AdminMovieHandler 影片管理 HTTP 层。
type AdminMovieHandler struct {
	movies *service.AdminMovieSvc
}

func NewAdminMovieHandler(movies *service.AdminMovieSvc) *AdminMovieHandler {
	return &AdminMovieHandler{movies: movies}
}

type movieRequest struct {
	Title           string  `json:"title" binding:"required"`
	CoverURL        string  `json:"cover_url"`
	TrailerURL      string  `json:"trailer_url"`
	Description     string  `json:"description"`
	DurationMinutes int     `json:"duration_minutes" binding:"required,gt=0"`
	Genre           string  `json:"genre"`
	ReleaseDate     string  `json:"release_date"` // 2006-01-02
	Rating          float64 `json:"rating"`
}

func (h *AdminMovieHandler) Create(c *gin.Context) {
	scope, ok := adminScopeFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "missing admin")
		return
	}
	var req movieRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	in, err := toMovieInput(req)
	if err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	movie, err := h.movies.Create(c.Request.Context(), scope.AdminID, in)
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, movie)
}

func (h *AdminMovieHandler) List(c *gin.Context) {
	movies, err := h.movies.List(c.Request.Context())
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, movies)
}

func (h *AdminMovieHandler) Update(c *gin.Context) {
	scope, ok := adminScopeFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "missing admin")
		return
	}
	movieID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, http.StatusBadRequest, "invalid movie id")
		return
	}
	var req movieRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	in, err := toMovieInput(req)
	if err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	movie, err := h.movies.Update(c.Request.Context(), scope.AdminID, movieID, in)
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, movie)
}

func (h *AdminMovieHandler) SetStatus(c *gin.Context) {
	scope, ok := adminScopeFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "missing admin")
		return
	}
	movieID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, http.StatusBadRequest, "invalid movie id")
		return
	}
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	status := domain.MovieStatus(req.Status)
	if status != domain.MovieOnSale && status != domain.MovieOffSale {
		resp.Fail(c, http.StatusBadRequest, "invalid status")
		return
	}
	if err := h.movies.SetStatus(c.Request.Context(), scope.AdminID, movieID, status); err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, gin.H{"id": movieID, "status": status})
}

func toMovieInput(req movieRequest) (service.MovieInput, error) {
	var releaseDate time.Time
	if req.ReleaseDate != "" {
		t, err := time.Parse("2006-01-02", req.ReleaseDate)
		if err != nil {
			return service.MovieInput{}, err
		}
		releaseDate = t
	}
	return service.MovieInput{
		Title:           req.Title,
		CoverURL:        req.CoverURL,
		TrailerURL:      req.TrailerURL,
		Description:     req.Description,
		DurationMinutes: req.DurationMinutes,
		Genre:           req.Genre,
		ReleaseDate:     releaseDate,
		Rating:          req.Rating,
	}, nil
}
