package middleware

import (
	"bank-sampah-backend/internal/model"
	"bank-sampah-backend/internal/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// AuditLogger automatically logs admin actions for auditable endpoints
func AuditLogger(auditRepo *repository.AuditRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Execute the handler first
		err := c.Next()

		// Only log successful state-changing operations
		if c.Response().StatusCode() >= 200 && c.Response().StatusCode() < 300 {
			// Only log POST, PUT, DELETE (not GET)
			method := c.Method()
			if method == "POST" || method == "PUT" || method == "DELETE" {
				go func() {
					adminIDStr, _ := c.Locals("admin_id").(string)
					var adminID *uuid.UUID
					if adminIDStr != "" {
						parsed, err := uuid.Parse(adminIDStr)
						if err == nil {
							adminID = &parsed
						}
					}

					auditRepo.Create(&model.AuditLog{
						AdminID:   adminID,
						Action:    method + " " + c.Path(),
						IPAddress: c.IP(),
						UserAgent: c.Get("User-Agent"),
					})
				}()
			}
		}

		return err
	}
}
