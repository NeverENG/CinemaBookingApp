package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/app/wire"
	"github.com/NeverENG/CinemaBookingApp/internal/infra/config"
	"github.com/NeverENG/CinemaBookingApp/internal/infra/database/postgres"
	"github.com/gin-gonic/gin"
)

// main 只做入口：解析参数 → 组装（wire）→ 启动。
func main() {
	migrate := flag.Bool("migrate", false, "apply sql/migrations and exit")
	flag.Parse()

	cfg := config.Load()
	if *migrate {
		db, err := wire.OpenDB(cfg)
		if err != nil {
			log.Fatalf("open db: %v", err)
		}
		if sqlDB, err := db.DB(); err == nil {
			defer sqlDB.Close()
		}
		if err := postgres.ApplyAllMigrations(db, "sql/migrations"); err != nil {
			log.Fatalf("migrate: %v", err)
		}
		if err := wire.EnsureBootstrap(cfg); err != nil {
			log.Fatalf("bootstrap: %v", err)
		}
		if err := postgres.ApplyMigrations(db, "sql/migrations/migrations010_seed_data.sql"); err != nil {
			log.Fatalf("seed: %v", err)
		}
		log.Println("migrations applied")
		return
	}

	gin.SetMode(cfg.GINMode)
	app, err := wire.NewApp(cfg)
	if err != nil {
		log.Fatalf("init app: %v", err)
	}

	server := &http.Server{
		Addr:              app.Addr,
		Handler:           app.Engine,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("listening on %s", app.Addr)
	jobCtx, cancelJobs := context.WithCancel(context.Background())
	jobDone := make(chan struct{})
	go func() {
		defer close(jobDone)
		app.Jobs.RunPeriodically(jobCtx, time.Minute)
	}()

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe()
	}()

	signalCtx, stopSignal := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignal()

	select {
	case <-signalCtx.Done():
		log.Println("shutdown signal received")
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server stopped: %v", err)
		}
	}

	cancelJobs()

	serverShutdownCtx, cancelServerShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	if err := server.Shutdown(serverShutdownCtx); err != nil {
		log.Printf("graceful server shutdown failed: %v", err)
		_ = server.Close()
	}
	cancelServerShutdown()

	jobShutdownCtx, cancelJobShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	select {
	case <-jobDone:
	case <-jobShutdownCtx.Done():
		log.Printf("job shutdown timed out: %v", jobShutdownCtx.Err())
	}
	cancelJobShutdown()

	select {
	case <-jobDone:
		if err := app.Close(); err != nil {
			log.Printf("database close failed: %v", err)
		}
	default:
		log.Println("database close skipped because jobs are still running")
	}
}
