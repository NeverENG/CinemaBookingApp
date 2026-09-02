package handler

import (
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/resp"
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/service"
	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/gin-gonic/gin"
	"strconv"
)

type AdminCinemaHandler struct{ cinemas *service.AdminCinemaSvc }

func NewAdminCinemaHandler(s *service.AdminCinemaSvc) *AdminCinemaHandler {
	return &AdminCinemaHandler{cinemas: s}
}

type cinemaRequest struct {
	Name      string  `json:"name" binding:"required"`
	City      string  `json:"city" binding:"required"`
	Address   string  `json:"address" binding:"required"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

func (h *AdminCinemaHandler) List(c *gin.Context) {
	scope, ok := adminScopeFrom(c)
	if !ok {
		resp.Fail(c, 401, "missing admin")
		return
	}
	v, e := h.cinemas.List(c, scope)
	if e != nil {
		resp.Error(c, e)
		return
	}
	resp.OK(c, v)
}
func (h *AdminCinemaHandler) Create(c *gin.Context) {
	scope, ok := adminScopeFrom(c)
	if !ok {
		resp.Fail(c, 401, "missing admin")
		return
	}
	var r cinemaRequest
	if e := c.ShouldBindJSON(&r); e != nil {
		resp.Fail(c, 400, e.Error())
		return
	}
	v, e := h.cinemas.Create(c, scope, service.CinemaInput{Name: r.Name, City: r.City, Address: r.Address, Longitude: r.Longitude, Latitude: r.Latitude})
	if e != nil {
		resp.Error(c, e)
		return
	}
	resp.OK(c, v)
}
func (h *AdminCinemaHandler) Update(c *gin.Context) {
	scope, ok := adminScopeFrom(c)
	if !ok {
		resp.Fail(c, 401, "missing admin")
		return
	}
	id, e := strconv.ParseInt(c.Param("id"), 10, 64)
	if e != nil {
		resp.Fail(c, 400, "invalid cinema id")
		return
	}
	var r cinemaRequest
	if e = c.ShouldBindJSON(&r); e != nil {
		resp.Fail(c, 400, e.Error())
		return
	}
	v, e := h.cinemas.Update(c, scope, id, service.CinemaInput{Name: r.Name, City: r.City, Address: r.Address, Longitude: r.Longitude, Latitude: r.Latitude})
	if e != nil {
		resp.Error(c, e)
		return
	}
	resp.OK(c, v)
}
func (h *AdminCinemaHandler) SetStatus(c *gin.Context) {
	scope, ok := adminScopeFrom(c)
	if !ok {
		resp.Fail(c, 401, "missing admin")
		return
	}
	id, e := strconv.ParseInt(c.Param("id"), 10, 64)
	if e != nil {
		resp.Fail(c, 400, "invalid cinema id")
		return
	}
	var r struct {
		Status string `json:"status" binding:"required"`
	}
	if e = c.ShouldBindJSON(&r); e != nil {
		resp.Fail(c, 400, e.Error())
		return
	}
	e = h.cinemas.SetStatus(c, scope, id, domain.CinemaStatus(r.Status))
	if e != nil {
		resp.Error(c, e)
		return
	}
	resp.OK(c, gin.H{"id": id, "status": r.Status})
}
