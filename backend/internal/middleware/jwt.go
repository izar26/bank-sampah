package middleware

import (
	"strings"

	"bank-sampah-backend/internal/service"
	"bank-sampah-backend/pkg/response"

	"github.com/gofiber/fiber/v2"
)

// JWTAuth validates JWT access tokens on protected routes
func JWTAuth(authSvc *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Unauthorized(c, "Token autentikasi diperlukan")
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return response.Unauthorized(c, "Format token tidak valid. Gunakan: Bearer <token>")
		}

		tokenString := parts[1]

		claims, err := authSvc.ValidateToken(tokenString)
		if err != nil {
			return response.Unauthorized(c, "Token tidak valid atau sudah kedaluwarsa")
		}

		if claims.Type != "access" {
			return response.Unauthorized(c, "Token bukan access token")
		}

		// Store admin info in context for downstream handlers
		c.Locals("admin_id", claims.AdminID)
		c.Locals("admin_username", claims.Username)

		return c.Next()
	}
}
