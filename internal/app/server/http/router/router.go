package router

import (
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/handler"
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/middleware"
	"github.com/gin-gonic/gin"
)

// New 注册路由。
func New(
	orderHandler *handler.OrderHandler,
	paymentHandler *handler.PaymentHandler,
	authHandler *handler.AuthHandler,
	authMw *middleware.AuthMiddleware,
) *gin.Engine {
	r := gin.Default()
	v1 := r.Group("/api/v1")
	{
		v1.POST("/auth/login", authHandler.UserLogin)
		v1.POST("/admin/auth/login", authHandler.AdminLogin)
		v1.POST("/orders", authMw.User(), orderHandler.Create)
		v1.POST("/payments", authMw.User(), paymentHandler.Create)
		// 模拟网关回调：不挂用户鉴权（真实场景由网关签名校验）
		v1.POST("/payments/mock-callback", paymentHandler.MockCallback)
	}
	return r
}
