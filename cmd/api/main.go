package main

import (
	"context"
	"log"
	"net/http"
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
	"github.com/smart-onestop/platform/internal/infrastructure/security"
	"github.com/smart-onestop/platform/internal/presentation"
	"github.com/smart-onestop/platform/internal/presentation/handler"
)

func main() {
	log.Println("================================================================================")
	log.Println(" Smart One-Stop Digital Citizen Hearing & Automated Escalation Platform (Go)")
	log.Println(" የ አንድ ማዕከል አገልግሎት - Clean Architecture / Modular Monolith")
	log.Println("================================================================================")

	// 1. Load 12-Factor Configuration
	cfg := config.LoadConfig()

	// 2. Initialize Database & Run Migrations
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Fatal: Database initialization failed: %v", err)
	}

	// 3. Seed Default Hierarchy, Users, and Demo Cases
	if err := database.SeedDatabase(db); err != nil {
		log.Printf("Warning: Database seeding encountered: %v", err)
	}

	// 4. Initialize Infrastructure Repositories
	caseRepo := repository.NewGormCaseRepository(db)
	structureRepo := repository.NewGormStructureRepository(db)
	userRepo := repository.NewGormUserRepository(db)
	citizenRepo := repository.NewGormCitizenRepository(db)
	hearingRepo := repository.NewGormHearingRepository(db)
	auditRepo := repository.NewGormAuditRepository(db)
	serviceTypeRepo := repository.NewGormServiceTypeRepository(db)

	// 5. Initialize Security & Redis
	jwtManager := security.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiration)
	redisClient := redis.NewClient(cfg.RedisURL)

	// 6. Initialize Application Services
	caseService := service.NewCaseService(
		caseRepo, citizenRepo, structureRepo, serviceTypeRepo, auditRepo, hearingRepo,
	)
	hearingService := service.NewHearingService(
		hearingRepo, caseRepo, auditRepo, userRepo,
	)
	slaService := service.NewSLAService(
		caseRepo, structureRepo, auditRepo,
	)
	authService := service.NewAuthService(
		userRepo, structureRepo, jwtManager,
	)
	analyticsService := service.NewAnalyticsService(
		caseRepo, structureRepo, auditRepo,
	)

	// 7. Start In-Process SLA Escalation Engine (Robfig Cron / Every 60s)
	// This ensures auto-escalation functions even in standalone mode without a separate worker container!
	c := cron.New(cron.WithSeconds())
	cronSpec := "@every 60s"
	_, err = c.AddFunc(cronSpec, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		locked, err := redisClient.AcquireLock(ctx, "sla_escalation_sweep", 45*time.Second)
		if err != nil || !locked {
			// Another worker instance holds the lock
			return
		}
		defer func() { _ = redisClient.ReleaseLock(ctx, "sla_escalation_sweep") }()

		count, err := slaService.SweepAndEscalate(ctx)
		if err != nil {
			log.Printf("[SLA Daemon] Sweep error: %v", err)
		} else if count > 0 {
			log.Printf("[SLA Daemon] Successfully swept and auto-escalated %d tickets", count)
		}
	})
	if err != nil {
		log.Printf("Warning: Failed to schedule SLA cron: %v", err)
	} else {
		c.Start()
		log.Printf("[SLA Daemon] Cron scheduled: checking expired tickets every 60s with distributed lock.")
	}
	defer c.Stop()

	// 8. Initialize Presentation Handlers & Router
	authHandler := handler.NewAuthHandler(authService)
	caseHandler := handler.NewCaseHandler(caseService)
	hearingHandler := handler.NewHearingHandler(hearingService, hearingRepo)
	structureHandler := handler.NewStructureHandler(structureRepo, serviceTypeRepo)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsService, slaService)

	router := presentation.SetupRouter(&presentation.RouterConfig{
		Cfg:              cfg,
		JWTManager:       jwtManager,
		AuthHandler:      authHandler,
		CaseHandler:      caseHandler,
		HearingHandler:   hearingHandler,
		StructureHandler: structureHandler,
		AnalyticsHandler: analyticsHandler,
	})

	// 9. Graceful HTTP Server Execution
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("[Server] Smart One-Stop API listening on http://localhost:%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for OS shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[Server] Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced shutdown: %v", err)
	}
	log.Println("[Server] Exited.")
}
