package handler

import (
	"net/http"

	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/resp"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HealthHandler 健康检查。
type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Check GET /healthz
func (h *HealthHandler) Check(c *gin.Context) {
	sqlDB, err := h.db.DB()
	if err != nil {
		resp.Fail(c, http.StatusServiceUnavailable, "db unavailable")
		return
	}
	if err := sqlDB.Ping(); err != nil {
		resp.Fail(c, http.StatusServiceUnavailable, "db unavailable")
		return
	}
	resp.OK(c, gin.H{"status": "ok"})
}
