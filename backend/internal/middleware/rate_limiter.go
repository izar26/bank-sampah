package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimiter returns a configured rate limiting middleware
// Default: 60 requests per minute per IP
func RateLimiter(maxRequests int, window time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        maxRequests,
		Expiration: window,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Use X-School-Key for SIMAK requests, IP for admin requests
			schoolKey := c.Get("X-School-Key")
			if schoolKey != "" {
				return "school:" + schoolKey
			}
			return "ip:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "Terlalu banyak request. Silakan coba lagi nanti.",
			})
		},
	})
}

// StrictRateLimiter for sensitive endpoints like login
func StrictRateLimiter() fiber.Handler {
	return RateLimiter(5, 1*time.Minute)
}

// APIRateLimiter for SIMAK API endpoints (per school)
func APIRateLimiter() fiber.Handler {
	return RateLimiter(30, 1*time.Minute)
}

// GeneralRateLimiter for admin dashboard endpoints
func GeneralRateLimiter() fiber.Handler {
	return RateLimiter(120, 1*time.Minute)
}
