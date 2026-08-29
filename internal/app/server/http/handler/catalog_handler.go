package handler

import (
	"net/http"
	"strconv"

	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/resp"
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/service"
	"github.com/gin-gonic/gin"
)

type CatalogHandler struct {
	catalog *service.CatalogSvc
}

func NewCatalogHandler(catalog *service.CatalogSvc) *CatalogHandler {
	return &CatalogHandler{catalog: catalog}
}

func (h *CatalogHandler) ListMovies(c *gin.Context) {
	views, err := h.catalog.ListMovies(c.Request.Context(), c.Query("keyword"), c.Query("status"))
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, views)
}

func (h *CatalogHandler) GetMovie(c *gin.Context) {
	movieID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || movieID <= 0 {
		resp.Fail(c, http.StatusBadRequest, "invalid movie id")
		return
	}
	view, err := h.catalog.GetMovie(c.Request.Context(), movieID)
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, view)
}

func (h *CatalogHandler) ListCinemas(c *gin.Context) {
	cinemas, err := h.catalog.ListCinemas(c.Request.Context(), c.Query("keyword"), c.Query("city"))
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, cinemas)
}
