package router

import (
	"time"

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
	refundHandler *handler.RefundHandler,
	dashboardHandler *handler.DashboardHandler,
	changeHandler *handler.ChangeHandler,
	healthHandler *handler.HealthHandler,
	couponHandler *handler.AdminCouponHandler,
	adminUserHandler *handler.AdminUserHandler,
	catalogHandler *handler.CatalogHandler,
	ticketHandler *handler.TicketHandler,
) *gin.Engine {
	r := gin.Default()
	_ = r.SetTrustedProxies(nil) // 不信任代理头，去除 gin 默认警告
	r.Use(middleware.CORS())
	r.GET("/healthz", healthHandler.Check)
	v1 := r.Group("/api/v1")
	v1.Use(middleware.RateLimit(middleware.NewTokenBucketLimiter(20, 40, 10*time.Minute), middleware.ClientIPKey))
	{
		v1.POST("/auth/login", authHandler.UserLogin)
		v1.POST("/auth/email-verification/request", authHandler.RequestRegistrationCode)
		v1.POST("/auth/register", authHandler.Register)
		v1.POST("/auth/password-reset/request", authHandler.RequestPasswordReset)
		v1.POST("/auth/password-reset/reset", authHandler.ResetPassword)
		v1.POST("/me/password", authMw.User(), authHandler.ChangePassword)
		v1.POST("/admin/auth/login", authHandler.AdminLogin)
		v1.GET("/sessions", userSessionHandler.List)
		v1.GET("/sessions/:id/seats", userSessionHandler.GetSeatMap)
		v1.GET("/home", homeHandler.Get)
		v1.GET("/movies", catalogHandler.ListMovies)
		v1.GET("/movies/:id", catalogHandler.GetMovie)
		v1.GET("/cinemas", catalogHandler.ListCinemas)
		v1.GET("/orders", authMw.User(), orderHandler.List)
		v1.POST("/orders", authMw.User(), orderHandler.Create)
		v1.GET("/orders/:order_no", authMw.User(), orderHandler.Get)
		v1.POST("/orders/:order_no/cancel", authMw.User(), orderHandler.Cancel)
		v1.POST("/orders/:order_no/refund", authMw.User(), refundHandler.Apply)
		v1.POST("/orders/:order_no/change", authMw.User(), changeHandler.Change)
		v1.POST("/payments", authMw.User(), paymentHandler.Create)
		v1.POST("/payments/mock-pay", authMw.User(), paymentHandler.MockPay)
		v1.GET("/payments/order/:order_no", authMw.User(), paymentHandler.GetByOrder)
		v1.GET("/me/points", authMw.User(), pointsHandler.Get)
		v1.POST("/me/points/exchange", authMw.User(), pointsHandler.Exchange)
		v1.GET("/coupons/redeemable", pointsHandler.ListRedeemable)
		v1.POST("/refunds/mock-callback", refundHandler.MockCallback)
		// 模拟网关回调：不挂用户鉴权（真实场景由网关签名校验）
		v1.POST("/payments/mock-callback", paymentHandler.MockCallback)

		admin := v1.Group("/admin")
		{
			admin.POST("/me/password", authMw.Admin(domain.RoleSuperAdmin, domain.RoleCinemaAdmin, domain.RoleFinance), authHandler.ChangeAdminPassword)

			admin.POST("/movies", authMw.Admin(domain.RoleSuperAdmin), movieHandler.Create)
			admin.GET("/movies", authMw.Admin(domain.RoleSuperAdmin, domain.RoleCinemaAdmin), movieHandler.List)
			admin.PATCH("/movies/:id", authMw.Admin(domain.RoleSuperAdmin), movieHandler.Update)
			admin.PATCH("/movies/:id/status", authMw.Admin(domain.RoleSuperAdmin), movieHandler.SetStatus)

			admin.POST("/halls", authMw.Admin(domain.RoleSuperAdmin, domain.RoleCinemaAdmin), hallHandler.Create)
			admin.GET("/halls", authMw.Admin(domain.RoleSuperAdmin, domain.RoleCinemaAdmin), hallHandler.List)
			admin.PATCH("/halls/:id", authMw.Admin(domain.RoleSuperAdmin, domain.RoleCinemaAdmin), hallHandler.Update)

			admin.POST("/sessions", authMw.Admin(domain.RoleSuperAdmin, domain.RoleCinemaAdmin), sessionHandler.Create)
			admin.PATCH("/sessions/:id/price", authMw.Admin(domain.RoleSuperAdmin, domain.RoleCinemaAdmin), sessionHandler.UpdatePrice)
			admin.POST("/sessions/:id/cancel", authMw.Admin(domain.RoleSuperAdmin, domain.RoleCinemaAdmin), sessionHandler.Cancel)

			admin.POST("/banners", authMw.Admin(domain.RoleSuperAdmin), bannerHandler.Create)
			admin.GET("/banners", authMw.Admin(domain.RoleSuperAdmin), bannerHandler.List)
			admin.PATCH("/banners/:id", authMw.Admin(domain.RoleSuperAdmin), bannerHandler.Update)
			admin.DELETE("/banners/:id", authMw.Admin(domain.RoleSuperAdmin), bannerHandler.Delete)

			admin.POST("/coupons/templates", authMw.Admin(domain.RoleSuperAdmin), couponHandler.CreateTemplate)
			admin.GET("/coupons/templates", authMw.Admin(domain.RoleSuperAdmin), couponHandler.ListTemplates)
			admin.PATCH("/coupons/templates/:id/status", authMw.Admin(domain.RoleSuperAdmin), couponHandler.SetTemplateStatus)
			admin.POST("/coupons/issue", authMw.Admin(domain.RoleSuperAdmin), couponHandler.IssueToUser)
			admin.GET("/admins", authMw.Admin(domain.RoleSuperAdmin), adminUserHandler.List)
			admin.POST("/admins", authMw.Admin(domain.RoleSuperAdmin), adminUserHandler.Create)
			admin.POST("/tickets/verify", authMw.Admin(domain.RoleSuperAdmin, domain.RoleCinemaAdmin), ticketHandler.Verify)

			dashboard := admin.Group("/dashboard", authMw.Admin(domain.RoleSuperAdmin, domain.RoleCinemaAdmin, domain.RoleFinance))
			dashboard.GET("/box-office", dashboardHandler.Trend)
			dashboard.GET("/box-office/summary", dashboardHandler.Summary)
			dashboard.GET("/box-office/by-movie", dashboardHandler.ByMovie)
			dashboard.GET("/box-office/by-cinema", dashboardHandler.ByCinema)
			dashboard.POST("/box-office/reconcile", authMw.Admin(domain.RoleSuperAdmin), dashboardHandler.Reconcile)
		}
	}
	return r
}
