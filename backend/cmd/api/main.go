package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"bank-sampah-backend/internal/config"
	"bank-sampah-backend/internal/database"
	"bank-sampah-backend/internal/repository"
	"bank-sampah-backend/internal/router"
	"bank-sampah-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// ── Load Configuration ──
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	// ── Connect to Database ──
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// ── Initialize Repositories ──
	adminRepo := repository.NewAdminRepository(db)
	schoolRepo := repository.NewSchoolRepository(db)
	siRepo := repository.NewSIRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	callbackRepo := repository.NewCallbackRepository(db)

	// ── Initialize Services ──
	authSvc := service.NewAuthService(adminRepo, cfg.JWT)
	callbackSvc := service.NewCallbackService(callbackRepo, cfg.Worker.CallbackMaxRetry)
	siSvc := service.NewSIService(siRepo, auditRepo, callbackRepo, schoolRepo)

	// ── Seed Default Admin ──
	if err := authSvc.SeedAdmin("admin", "admin@banksampah.id", "password123"); err != nil {
		log.Printf("⚠️ Failed to seed admin: %v", err)
	} else {
		log.Println("✅ Default admin seeded (username: admin, password: password123)")
	}

	// ── Start Background Workers ──
	workerPool := service.NewWorkerPool(callbackSvc, cfg.Worker.PoolSize)
	workerPool.Start()

	// ── Initialize Fiber App ──
	app := fiber.New(fiber.Config{
		AppName:       "Bank Sampah API v1.0.0",
		BodyLimit:     10 * 1024 * 1024, // 10MB max body
		ServerHeader:  "BankSampah",
		StrictRouting: false,
		CaseSensitive: false,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		},
	})

	// ── Setup Routes ──
	router.Setup(app, authSvc, siSvc, schoolRepo, callbackRepo, auditRepo, cfg.HMAC.TimestampTolerance)

	// ── Graceful Shutdown ──
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("\n🛑 Shutting down gracefully...")
		workerPool.Stop()

		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}

		app.Shutdown()
	}()

	// ── Start Server ──
	addr := fmt.Sprintf(":%s", cfg.App.Port)
	log.Printf("🚀 Bank Sampah API starting on http://localhost%s", addr)
	log.Printf("📖 Health check: http://localhost%s/health", addr)
	log.Printf("🔑 Default login: admin / password123")

	if err := app.Listen(addr); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}
