package router

import (
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/handler"
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/middleware"
	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/gin-gonic/gin"
)

// New 注册路由。
func New(
	orderHandler *handler.OrderHandler,
	paymentHandler *handler.PaymentHandler,
	authHandler *handler.AuthHandler,
	authMw *middleware.AuthMiddleware,
	movieHandler *handler.AdminMovieHandler,
	hallHandler *handler.AdminHallHandler,
	sessionHandler *handler.AdminSessionHandler,
	userSessionHandler *handler.SessionHandler,
	homeHandler *handler.HomeHandler,
	bannerHandler *handler.AdminBannerHandler,
	pointsHandler *handler.PointsHandler,
) *gin.Engine {
	r := gin.Default()
	v1 := r.Group("/api/v1")
	{
		v1.POST("/auth/login", authHandler.UserLogin)
		v1.POST("/admin/auth/login", authHandler.AdminLogin)
		v1.GET("/sessions", userSessionHandler.List)
		v1.GET("/sessions/:id/seats", userSessionHandler.GetSeatMap)
		v1.GET("/home", homeHandler.Get)
		v1.POST("/orders", authMw.User(), orderHandler.Create)
		v1.GET("/orders/:order_no", authMw.User(), orderHandler.Get)
		v1.POST("/payments", authMw.User(), paymentHandler.Create)
		v1.POST("/payments/mock-pay", authMw.User(), paymentHandler.MockPay)
		v1.GET("/payments/order/:order_no", authMw.User(), paymentHandler.GetByOrder)
		v1.GET("/me/points", authMw.User(), pointsHandler.Get)
		// 模拟网关回调：不挂用户鉴权（真实场景由网关签名校验）
		v1.POST("/payments/mock-callback", paymentHandler.MockCallback)

		admin := v1.Group("/admin")
		{
			admin.POST("/movies", authMw.Admin(domain.RoleSuperAdmin, domain.RoleCinemaAdmin), movieHandler.Create)
			admin.GET("/movies", authMw.Admin(domain.RoleSuperAdmin, domain.RoleCinemaAdmin), movieHandler.List)
			admin.PATCH("/movies/:id", authMw.Admin(domain.RoleSuperAdmin, domain.RoleCinemaAdmin), movieHandler.Update)
			admin.PATCH("/movies/:id/status", authMw.Admin(domain.RoleSuperAdmin, domain.RoleCinemaAdmin), movieHandler.SetStatus)

			admin.POST("/halls", authMw.Admin(domain.RoleSuperAdmin, domain.RoleCinemaAdmin), hallHandler.Create)
			admin.GET("/halls", authMw.Admin(domain.RoleSuperAdmin, domain.RoleCinemaAdmin), hallHandler.List)
			admin.PATCH("/halls/:id", authMw.Admin(domain.RoleSuperAdmin, domain.RoleCinemaAdmin), hallHandler.Update)

			admin.POST("/sessions", authMw.Admin(domain.RoleSuperAdmin, domain.RoleCinemaAdmin), sessionHandler.Create)
			admin.PATCH("/sessions/:id/price", authMw.Admin(domain.RoleSuperAdmin, domain.RoleCinemaAdmin, domain.RoleFinance), sessionHandler.UpdatePrice)
			admin.POST("/sessions/:id/cancel", authMw.Admin(domain.RoleSuperAdmin, domain.RoleCinemaAdmin), sessionHandler.Cancel)

			admin.POST("/banners", authMw.Admin(domain.RoleSuperAdmin, domain.RoleCinemaAdmin), bannerHandler.Create)
			admin.GET("/banners", authMw.Admin(domain.RoleSuperAdmin, domain.RoleCinemaAdmin), bannerHandler.List)
			admin.PATCH("/banners/:id", authMw.Admin(domain.RoleSuperAdmin, domain.RoleCinemaAdmin), bannerHandler.Update)
			admin.DELETE("/banners/:id", authMw.Admin(domain.RoleSuperAdmin, domain.RoleCinemaAdmin), bannerHandler.Delete)
		}
	}
	return r
}
