package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/app/wire"
	"github.com/NeverENG/CinemaBookingApp/internal/infra/database/postgres"
)

// main 只做入口：解析参数 → 组装（wire）→ 启动。
func main() {
	migrate := flag.Bool("migrate", false, "apply sql/migrations and exit")
	flag.Parse()

	app, err := wire.NewApp()
	if err != nil {
		log.Fatalf("init app: %v", err)
	}

	if *migrate {
		if err := postgres.ApplyAllMigrations(app.DB, "sql/migrations"); err != nil {
			log.Fatalf("migrate: %v", err)
		}
		log.Println("migrations applied")
		return
	}

	log.Printf("listening on %s", app.Addr)
	jobCtx, cancelJobs := context.WithCancel(context.Background())
	defer cancelJobs()
	go app.Jobs.RunPeriodically(jobCtx, time.Minute)
	if err := app.Engine.Run(app.Addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}
