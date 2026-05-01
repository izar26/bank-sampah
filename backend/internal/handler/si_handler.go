package handler

import (
	"strconv"

	"bank-sampah-backend/internal/service"
	"bank-sampah-backend/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type SIHandler struct {
	siSvc *service.SIService
}

func NewSIHandler(siSvc *service.SIService) *SIHandler {
	return &SIHandler{siSvc: siSvc}
}

// SubmitSI receives an SI from SIMAK (HMAC-protected)
// POST /api/v1/si/submit
func (h *SIHandler) SubmitSI(c *fiber.Ctx) error {
	schoolIDStr, _ := c.Locals("school_id").(string)
	schoolID, err := uuid.Parse(schoolIDStr)
	if err != nil {
		return response.BadRequest(c, "School ID tidak valid")
	}

	var req service.SubmitSIRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Format request tidak valid")
	}

	if req.SINumber == "" {
		return response.BadRequest(c, "Nomor SI wajib diisi")
	}
	if req.IdempotencyKey == "" {
		return response.BadRequest(c, "Idempotency key wajib diisi")
	}

	doc, err := h.siSvc.SubmitSI(schoolID, &req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Accepted(c, "Surat Instruksi berhasil diterima dan sedang diproses", fiber.Map{
		"si_id":       doc.ID,
		"si_number":   doc.SINumber,
		"status":      doc.Status,
		"total_items":  doc.TotalItems,
		"total_amount": doc.TotalAmount,
	})
}

// GetSIStatus checks the status of an SI (HMAC-protected, for SIMAK)
// GET /api/v1/si/:id/status
func (h *SIHandler) GetSIStatus(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(c, "ID tidak valid")
	}

	doc, err := h.siSvc.GetDocumentByID(id)
	if err != nil {
		return response.NotFound(c, "Dokumen SI tidak ditemukan")
	}

	return response.Success(c, "Status SI", fiber.Map{
		"si_id":       doc.ID,
		"si_number":   doc.SINumber,
		"status":      doc.Status,
		"total_items":  doc.TotalItems,
		"total_amount": doc.TotalAmount,
		"verified_at":  doc.VerifiedAt,
		"approved_at":  doc.ApprovedAt,
		"disbursed_at": doc.DisbursedAt,
	})
}

// ListSI returns paginated SI documents for the admin dashboard
// GET /api/v1/si
func (h *SIHandler) ListSI(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "15"))
	status := c.Query("status")

	var schoolID *uuid.UUID
	if schoolIDStr := c.Query("school_id"); schoolIDStr != "" {
		if parsed, err := uuid.Parse(schoolIDStr); err == nil {
			schoolID = &parsed
		}
	}

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 15
	}

	docs, total, err := h.siSvc.ListDocuments(page, perPage, status, schoolID)
	if err != nil {
		return response.InternalError(c, "Gagal mengambil data SI")
	}

	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	return response.Paginated(c, docs, &response.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

// GetSIDetail returns full SI document with items
// GET /api/v1/si/:id
func (h *SIHandler) GetSIDetail(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "ID tidak valid")
	}

	doc, err := h.siSvc.GetDocumentByID(id)
	if err != nil {
		return response.NotFound(c, "Dokumen SI tidak ditemukan")
	}

	return response.Success(c, "Detail SI", doc)
}

// VerifySI transitions SI to VERIFIED
// PUT /api/v1/si/:id/verify
func (h *SIHandler) VerifySI(c *fiber.Ctx) error {
	siID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "ID tidak valid")
	}

	adminID, err := uuid.Parse(c.Locals("admin_id").(string))
	if err != nil {
		return response.Unauthorized(c, "Admin ID tidak valid")
	}

	if err := h.siSvc.VerifySI(siID, adminID, c.IP()); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Success(c, "SI berhasil diverifikasi", nil)
}

// ApproveSI transitions SI to APPROVED
// PUT /api/v1/si/:id/approve
func (h *SIHandler) ApproveSI(c *fiber.Ctx) error {
	siID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "ID tidak valid")
	}

	adminID, err := uuid.Parse(c.Locals("admin_id").(string))
	if err != nil {
		return response.Unauthorized(c, "Admin ID tidak valid")
	}

	if err := h.siSvc.ApproveSI(siID, adminID, c.IP()); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Success(c, "SI berhasil di-approve", nil)
}

// DisburseSI transitions SI to DISBURSED
// PUT /api/v1/si/:id/disburse
func (h *SIHandler) DisburseSI(c *fiber.Ctx) error {
	siID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "ID tidak valid")
	}

	adminID, err := uuid.Parse(c.Locals("admin_id").(string))
	if err != nil {
		return response.Unauthorized(c, "Admin ID tidak valid")
	}

	if err := h.siSvc.DisburseSI(siID, adminID, c.IP()); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Success(c, "SI berhasil dicairkan", nil)
}

// RejectSI transitions SI to REJECTED
// PUT /api/v1/si/:id/reject
func (h *SIHandler) RejectSI(c *fiber.Ctx) error {
	siID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "ID tidak valid")
	}

	adminID, err := uuid.Parse(c.Locals("admin_id").(string))
	if err != nil {
		return response.Unauthorized(c, "Admin ID tidak valid")
	}

	type RejectRequest struct {
		Reason string `json:"reason"`
	}
	var req RejectRequest
	if err := c.BodyParser(&req); err != nil {
		req.Reason = "Ditolak oleh admin"
	}

	if err := h.siSvc.RejectSI(siID, adminID, req.Reason, c.IP()); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Success(c, "SI berhasil ditolak", nil)
}
