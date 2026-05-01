package handler

import (
	"strconv"

	"bank-sampah-backend/internal/repository"
	"bank-sampah-backend/internal/service"
	"bank-sampah-backend/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type DashboardHandler struct {
	siSvc     *service.SIService
	auditRepo *repository.AuditRepository
}

func NewDashboardHandler(siSvc *service.SIService, auditRepo *repository.AuditRepository) *DashboardHandler {
	return &DashboardHandler{
		siSvc:     siSvc,
		auditRepo: auditRepo,
	}
}

// GetDashboard returns aggregated statistics
// GET /api/v1/dashboard
func (h *DashboardHandler) GetDashboard(c *fiber.Ctx) error {
	stats, err := h.siSvc.GetDashboard()
	if err != nil {
		return response.InternalError(c, "Gagal mengambil data dashboard")
	}
	return response.Success(c, "Data dashboard", stats)
}

// GetAuditLogs returns paginated audit trail
// GET /api/v1/audit-logs
func (h *DashboardHandler) GetAuditLogs(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	var adminID *uuid.UUID
	if adminIDStr := c.Query("admin_id"); adminIDStr != "" {
		if parsed, err := uuid.Parse(adminIDStr); err == nil {
			adminID = &parsed
		}
	}

	logs, total, err := h.auditRepo.FindAll(page, perPage, adminID)
	if err != nil {
		return response.InternalError(c, "Gagal mengambil audit log")
	}

	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	return response.Paginated(c, logs, &response.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}
