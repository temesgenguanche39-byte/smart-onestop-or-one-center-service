package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/smart-onestop/platform/internal/application/service"
	"github.com/smart-onestop/platform/internal/infrastructure/config"
	"github.com/smart-onestop/platform/internal/infrastructure/database"
	"github.com/smart-onestop/platform/internal/infrastructure/redis"
	"github.com/smart-onestop/platform/internal/infrastructure/repository"
)

func main() {
	log.Println("================================================================================")
	log.Println(" Smart One-Stop Platform: Automated SLA Worker Daemon")
	log.Println(" Race-Condition-Free Hierarchical Escalation Engine")
	log.Println("================================================================================")

	cfg := config.LoadConfig()

	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Fatal: Database initialization failed: %v", err)
	}

	caseRepo := repository.NewGormCaseRepository(db)
	structureRepo := repository.NewGormStructureRepository(db)
	auditRepo := repository.NewGormAuditRepository(db)
	holidayRepo := repository.NewGormHolidayRepository(db)
	_ = service.NewCalendarService(holidayRepo, auditRepo)

	redisClient := redis.NewClient(cfg.RedisURL)
	slaService := service.NewSLAService(caseRepo, structureRepo, auditRepo)

	c := cron.New(cron.WithSeconds())
	cronSpec := "@every 60s"

	log.Printf("[Worker] Starting periodic SLA sweep scheduler (%s)...", cronSpec)

	_, err = c.AddFunc(cronSpec, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()

		log.Println("[Worker] Running scheduled SLA breach detection sweep...")

		// Acquire distributed lock to prevent multi-worker concurrency
		locked, err := redisClient.AcquireLock(ctx, "sla_worker_distributed_lock", 50*time.Second)
		if err != nil {
			log.Printf("[Worker] Redis lock error: %v", err)
			return
		}
		if !locked {
			log.Println("[Worker] Lock held by another worker instance; skipping this sweep.")
			return
		}
		defer func() {
			_ = redisClient.ReleaseLock(ctx, "sla_worker_distributed_lock")
		}()

		count, err := slaService.SweepAndEscalate(ctx)
		if err != nil {
			log.Printf("[Worker] Error during SLA sweep: %v", err)
			return
		}

		if count > 0 {
			log.Printf("[Worker] Successfully auto-escalated %d breached tickets to higher municipal tiers.", count)
		} else {
			log.Println("[Worker] SLA check completed. No new breached cases found.")
		}
	})

	if err != nil {
		log.Fatalf("Fatal: Failed to schedule cron task: %v", err)
	}

	c.Start()
	defer c.Stop()

	// Wait for termination
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[Worker] Stopping SLA Worker Daemon gracefully...")
}
