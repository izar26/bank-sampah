package router

import (
	"bank-sampah-backend/internal/handler"
	"bank-sampah-backend/internal/middleware"
	"bank-sampah-backend/internal/repository"
	"bank-sampah-backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// Setup configures all routes and middleware
func Setup(
	app *fiber.App,
	authSvc *service.AuthService,
	siSvc *service.SIService,
	schoolRepo *repository.SchoolRepository,
	callbackRepo *repository.CallbackRepository,
	auditRepo *repository.AuditRepository,
	hmacTolerance int,
) {
	// ── Global Middleware ──
	app.Use(recover.New())
	app.Use(fiberlogger.New(fiberlogger.Config{
		Format:     "${time} | ${status} | ${latency} | ${ip} | ${method} ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-School-Key, X-Signature, X-Timestamp, X-Nonce",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// ── Handlers ──
	authHandler := handler.NewAuthHandler(authSvc)
	siHandler := handler.NewSIHandler(siSvc)
	schoolHandler := handler.NewSchoolHandler(schoolRepo)
	dashboardHandler := handler.NewDashboardHandler(siSvc, auditRepo)

	// ── Health Check ──
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Bank Sampah API is running 🚀",
			"version": "1.0.0",
		})
	})

	// ── API v1 ──
	v1 := app.Group("/api/v1")

	// ── Public: Auth (rate limited) ──
	auth := v1.Group("/auth")
	auth.Post("/login", middleware.StrictRateLimiter(), authHandler.Login)
	auth.Post("/refresh", middleware.StrictRateLimiter(), authHandler.RefreshToken)

	// ── SIMAK Endpoints (HMAC-protected) ──
	// Note: We don't apply middleware to the Group itself to avoid intercepting Admin /si routes
	simak := v1.Group("/si")
	simak.Post("/submit", 
		middleware.APIRateLimiter(), 
		middleware.HMACAuth(schoolRepo, callbackRepo, hmacTolerance), 
		siHandler.SubmitSI,
	)
	simak.Get("/:id/status", 
		middleware.APIRateLimiter(), 
		middleware.HMACAuth(schoolRepo, callbackRepo, hmacTolerance), 
		siHandler.GetSIStatus,
	)

	// ── Admin Endpoints (JWT-protected) ──
	admin := v1.Group("",
		middleware.GeneralRateLimiter(),
		middleware.JWTAuth(authSvc),
		middleware.AuditLogger(auditRepo),
	)

	// Auth
	admin.Get("/auth/me", authHandler.Me)

	// Dashboard
	admin.Get("/dashboard", dashboardHandler.GetDashboard)

	// SI Management
	admin.Get("/si", siHandler.ListSI)
	admin.Get("/si/:id", siHandler.GetSIDetail)
	admin.Put("/si/:id/verify", siHandler.VerifySI)
	admin.Put("/si/:id/approve", siHandler.ApproveSI)
	admin.Put("/si/:id/disburse", siHandler.DisburseSI)
	admin.Put("/si/:id/reject", siHandler.RejectSI)

	// School Management
	admin.Get("/schools", schoolHandler.ListSchools)
	admin.Get("/schools/:id", schoolHandler.GetSchool)
	admin.Post("/schools", schoolHandler.CreateSchool)
	admin.Put("/schools/:id", schoolHandler.UpdateSchool)
	admin.Delete("/schools/:id", schoolHandler.DeleteSchool)
	admin.Post("/schools/:id/regenerate", schoolHandler.RegenerateCredentials)

	// Audit Logs
	admin.Get("/audit-logs", dashboardHandler.GetAuditLogs)
}
