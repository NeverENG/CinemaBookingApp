// Package wire 手写依赖注入：全项目唯一知道「谁实现谁」的地方。
package wire

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/handler"
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/middleware"
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/router"
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/service"
	"github.com/NeverENG/CinemaBookingApp/internal/infra/database/postgres"
	"github.com/NeverENG/CinemaBookingApp/internal/pkg/jwt"
	"github.com/gin-gonic/gin"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// App 组装完成的运行时对象，入口只需要拿到它启动。
type App struct {
	Engine *gin.Engine
	DB     *gorm.DB
	Addr   string
}

// NewApp 按依赖方向手工组装：config → db → repos → services → handlers → router。
func NewApp() (*App, error) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=cinema port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	}
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	db, err := gorm.Open(pgdriver.Open(dsn), &gorm.Config{})
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

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret"
	}
	tokens := jwt.New(jwtSecret, 24*time.Hour)

	orderSvc := service.NewOrderSvc(txm, userRepo, sessionRepo, seatRepo, seatLockRepo, couponRepo, orderRepo)
	paymentSvc := service.NewPaymentSvc(txm, paymentRepo, callbackRepo, orderRepo, seatLockRepo, couponRepo)
	authSvc := service.NewAuthSvc(userRepo, adminRepo, roleRepo, tokens)
	if err := authSvc.EnsureDefaultAdmin(context.Background()); err != nil {
		log.Printf("ensure default admin: %v", err)
	}

	movieSvc := service.NewAdminMovieSvc(movieRepo, operationLogRepo)
	hallSvc := service.NewAdminHallSvc(hallRepo, seatRepo, operationLogRepo)
	sessionSvc := service.NewAdminSessionSvc(sessionRepo, movieRepo, hallRepo, seatLockRepo, orderRepo, couponRepo, operationLogRepo)

	orderHandler := handler.NewOrderHandler(orderSvc)
	paymentHandler := handler.NewPaymentHandler(paymentSvc)
	authHandler := handler.NewAuthHandler(authSvc)
	movieHandler := handler.NewAdminMovieHandler(movieSvc)
	hallHandler := handler.NewAdminHallHandler(hallSvc)
	sessionHandler := handler.NewAdminSessionHandler(sessionSvc)
	authMw := middleware.NewAuthMiddleware(tokens)

	engine := router.New(orderHandler, paymentHandler, authHandler, authMw, movieHandler, hallHandler, sessionHandler)
	return &App{Engine: engine, DB: db, Addr: addr}, nil
}
