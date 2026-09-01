package handler

import (
	"net/http"
	"strconv"

	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/resp"
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/service"
	"github.com/gin-gonic/gin"
)

// AdminBannerHandler banner 管理。
type AdminBannerHandler struct {
	banners *service.AdminBannerSvc
}

func NewAdminBannerHandler(banners *service.AdminBannerSvc) *AdminBannerHandler {
	return &AdminBannerHandler{banners: banners}
}

type bannerRequest struct {
	Title    string `json:"title" binding:"required"`
	ImageURL string `json:"image_url" binding:"required"`
	Sort     int    `json:"sort"`
	Enabled  *bool  `json:"enabled"`
}

func (h *AdminBannerHandler) Create(c *gin.Context) {
	scope, ok := adminScopeFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "missing admin")
		return
	}
	var req bannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	banner, err := h.banners.Create(c.Request.Context(), scope, toBannerInput(req))
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, banner)
}

func (h *AdminBannerHandler) List(c *gin.Context) {
	scope, ok := adminScopeFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "missing admin")
		return
	}
	banners, err := h.banners.List(c.Request.Context(), scope)
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, banners)
}

func (h *AdminBannerHandler) Update(c *gin.Context) {
	scope, ok := adminScopeFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "missing admin")
		return
	}
	bannerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, http.StatusBadRequest, "invalid banner id")
		return
	}
	var req bannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	banner, err := h.banners.Update(c.Request.Context(), scope, bannerID, toBannerInput(req))
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, banner)
}

func (h *AdminBannerHandler) Delete(c *gin.Context) {
	scope, ok := adminScopeFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "missing admin")
		return
	}
	bannerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, http.StatusBadRequest, "invalid banner id")
		return
	}
	if err := h.banners.Delete(c.Request.Context(), scope, bannerID); err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, gin.H{"id": bannerID})
}

func toBannerInput(req bannerRequest) service.BannerInput {
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	return service.BannerInput{
		Title:    req.Title,
		ImageURL: req.ImageURL,
		Sort:     req.Sort,
		Enabled:  enabled,
	}
}
