package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/app/wire"
	"github.com/NeverENG/CinemaBookingApp/internal/infra/config"
	"github.com/NeverENG/CinemaBookingApp/internal/infra/database/postgres"
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

	app, err := wire.NewApp(cfg)
	if err != nil {
		log.Fatalf("init app: %v", err)
	}

	log.Printf("listening on %s", app.Addr)
	jobCtx, cancelJobs := context.WithCancel(context.Background())
	defer cancelJobs()
	go app.Jobs.RunPeriodically(jobCtx, time.Minute)
	if err := app.Engine.Run(app.Addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}
