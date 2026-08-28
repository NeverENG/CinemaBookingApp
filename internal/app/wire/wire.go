// Package wire 手写依赖注入：全项目唯一知道「谁实现谁」的地方。
package wire

import (
	"context"
	"log"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/app/job"
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/handler"
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/middleware"
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/router"
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/service"
	"github.com/NeverENG/CinemaBookingApp/internal/infra/config"
	"github.com/NeverENG/CinemaBookingApp/internal/infra/database/postgres"
	"github.com/NeverENG/CinemaBookingApp/internal/pkg/jwt"
	"github.com/NeverENG/CinemaBookingApp/internal/pkg/mailer"
	"github.com/gin-gonic/gin"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// App 组装完成的运行时对象，入口只需要拿到它启动。
type App struct {
	Engine *gin.Engine
	DB     *gorm.DB
	Addr   string
	Jobs   *job.Runner
}

// OpenDB 按配置打开数据库连接。
func OpenDB(cfg config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(pgdriver.Open(cfg.DB.DSN()), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

// NewApp 按依赖方向手工组装：config → db → repos → services → handlers → router。
func NewApp(cfg config.Config) (*App, error) {
	db, err := OpenDB(cfg)
	if err != nil {
		return nil, err
	}
	log.Printf("database connected")

	pg := postgres.NewDB(db)
	txm := postgres.NewTxManager(pg)

	userRepo := postgres.NewUserRepo(pg)
	sessionRepo := postgres.NewSessionRepo(pg)
	seatRepo := postgres.NewSeatRepo(pg)
	seatLockRepo := postgres.NewSeatLockRepo(pg)
	couponRepo := postgres.NewUserCouponRepo(pg)
	orderRepo := postgres.NewOrderRepo(pg)
	paymentRepo := postgres.NewPaymentRepo(pg)
	callbackRepo := postgres.NewPaymentCallbackRepo(pg)
	adminRepo := postgres.NewAdminRepo(pg)
	roleRepo := postgres.NewRoleRepo(pg)
	movieRepo := postgres.NewMovieRepo(pg)
	hallRepo := postgres.NewHallRepo(pg)
	operationLogRepo := postgres.NewOperationLogRepo(pg)
	refundRepo := postgres.NewRefundRepo(pg)
	bannerRepo := postgres.NewBannerRepo(pg)
	pointsRepo := postgres.NewPointsRepo(pg)
	boxOfficeRepo := postgres.NewBoxOfficeRepo(pg)
	passwordResetRepo := postgres.NewPasswordResetRepo(pg)
	membershipRepo := postgres.NewMembershipRepo(pg)
	mailerCfg := mailer.FromEnv()

	tokens := jwt.New(cfg.JWTSecret, cfg.JWTExpire)

	orderSvc := service.NewOrderSvc(txm, userRepo, sessionRepo, seatRepo, seatLockRepo, couponRepo, orderRepo)
	paymentSvc := service.NewPaymentSvc(txm, paymentRepo, callbackRepo, orderRepo, seatLockRepo, couponRepo, pointsRepo, boxOfficeRepo, membershipRepo)
	authSvc := service.NewAuthSvc(
		userRepo, adminRepo, roleRepo, tokens,
		service.Bootstrap{
			AdminUsername: cfg.Admin.Username,
			AdminPassword: cfg.Admin.Password,
			DemoUsername:  cfg.Demo.Username,
			DemoPassword:  cfg.Demo.Password,
		},
		passwordResetRepo, membershipRepo, mailerCfg, mailerCfg.Enabled(),
	)
	if err := authSvc.EnsureDefaultAdmin(context.Background()); err != nil {
		log.Printf("ensure default admin: %v", err)
	}
	if err := authSvc.EnsureDemoUser(context.Background()); err != nil {
		log.Printf("ensure demo user: %v", err)
	}

	movieSvc := service.NewAdminMovieSvc(movieRepo, operationLogRepo)
	hallSvc := service.NewAdminHallSvc(hallRepo, seatRepo, operationLogRepo)
	sessionSvc := service.NewAdminSessionSvc(sessionRepo, movieRepo, hallRepo, seatLockRepo, orderRepo, couponRepo, refundRepo, paymentRepo, pointsRepo, boxOfficeRepo, operationLogRepo)
	seatMapSvc := service.NewSeatMapSvc(sessionRepo, seatRepo, seatLockRepo, movieRepo, hallRepo)
	homeSvc := service.NewHomeSvc(bannerRepo, movieRepo, orderRepo)
	bannerSvc := service.NewAdminBannerSvc(bannerRepo, operationLogRepo)
	pointsSvc := service.NewPointsSvc(txm, pointsRepo, couponRepo)
	refundSvc := service.NewRefundSvc(txm, orderRepo, refundRepo, paymentRepo, seatLockRepo, pointsRepo, sessionRepo, boxOfficeRepo)
	boxOfficeSvc := service.NewBoxOfficeSvc(boxOfficeRepo)
	changeSvc := service.NewChangeTicketSvc(orderRepo, sessionRepo, orderSvc, paymentSvc, refundSvc)

	orderHandler := handler.NewOrderHandler(orderSvc)
	paymentHandler := handler.NewPaymentHandler(paymentSvc)
	authHandler := handler.NewAuthHandler(authSvc)
	movieHandler := handler.NewAdminMovieHandler(movieSvc)
	hallHandler := handler.NewAdminHallHandler(hallSvc)
	sessionHandler := handler.NewAdminSessionHandler(sessionSvc)
	userSessionHandler := handler.NewSessionHandler(seatMapSvc)
	homeHandler := handler.NewHomeHandler(homeSvc)
	bannerHandler := handler.NewAdminBannerHandler(bannerSvc)
	pointsHandler := handler.NewPointsHandler(pointsSvc)
	refundHandler := handler.NewRefundHandler(refundSvc)
	dashboardHandler := handler.NewDashboardHandler(boxOfficeSvc)
	changeHandler := handler.NewChangeHandler(changeSvc)
	healthHandler := handler.NewHealthHandler(db)
	authMw := middleware.NewAuthMiddleware(tokens)

	engine := router.New(orderHandler, paymentHandler, authHandler, authMw, movieHandler, hallHandler, sessionHandler, userSessionHandler, homeHandler, bannerHandler, pointsHandler, refundHandler, dashboardHandler, changeHandler, healthHandler)

	runner := job.NewRunner()
	runner.Add("order_timeout", func(ctx context.Context) error {
		_, err := orderSvc.ExpireOverdueOrders(ctx, time.Now())
		return err
	})
	runner.Add("payment_callback_retry", func(ctx context.Context) error {
		_, err := paymentSvc.RetryCallbacks(ctx, 50)
		return err
	})
	runner.Add("payment_timeout", func(ctx context.Context) error {
		_, err := paymentSvc.CloseExpiredPending(ctx, 100)
		return err
	})
	runner.Add("box_office_reconcile", func(ctx context.Context) error {
		return boxOfficeSvc.Reconcile(ctx)
	})

	return &App{Engine: engine, DB: db, Addr: cfg.HTTPAddr, Jobs: runner}, nil
}

// EnsureBootstrap 迁移后引导默认管理员与演示用户（迁移需要先于账号存在）。
func EnsureBootstrap(cfg config.Config) error {
	db, err := OpenDB(cfg)
	if err != nil {
		return err
	}
	pg := postgres.NewDB(db)
	userRepo := postgres.NewUserRepo(pg)
	adminRepo := postgres.NewAdminRepo(pg)
	roleRepo := postgres.NewRoleRepo(pg)
	tokens := jwt.New(cfg.JWTSecret, cfg.JWTExpire)
	authSvc := service.NewAuthSvc(
		userRepo, adminRepo, roleRepo, tokens,
		service.Bootstrap{
			AdminUsername: cfg.Admin.Username,
			AdminPassword: cfg.Admin.Password,
			DemoUsername:  cfg.Demo.Username,
			DemoPassword:  cfg.Demo.Password,
		},
		postgres.NewPasswordResetRepo(pg), postgres.NewMembershipRepo(pg), mailer.FromEnv(), false,
	)
	if err := authSvc.EnsureDefaultAdmin(context.Background()); err != nil {
		return err
	}
	return authSvc.EnsureDemoUser(context.Background())
}
