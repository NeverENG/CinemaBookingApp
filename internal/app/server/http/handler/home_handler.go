package handler

import (
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/resp"
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/service"
	"github.com/gin-gonic/gin"
)

// HomeHandler 首页公开接口。
type HomeHandler struct {
	home *service.HomeSvc
}

func NewHomeHandler(home *service.HomeSvc) *HomeHandler {
	return &HomeHandler{home: home}
}

// Get GET /api/v1/home
func (h *HomeHandler) Get(c *gin.Context) {
	view, err := h.home.GetHome(c.Request.Context())
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, view)
}
