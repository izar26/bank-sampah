package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"bank-sampah-backend/internal/model"
	"bank-sampah-backend/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SIService struct {
	siRepo       *repository.SIRepository
	auditRepo    *repository.AuditRepository
	callbackRepo *repository.CallbackRepository
	schoolRepo   *repository.SchoolRepository
}

func NewSIService(
	siRepo *repository.SIRepository,
	auditRepo *repository.AuditRepository,
	callbackRepo *repository.CallbackRepository,
	schoolRepo *repository.SchoolRepository,
) *SIService {
	return &SIService{
		siRepo:       siRepo,
		auditRepo:    auditRepo,
		callbackRepo: callbackRepo,
		schoolRepo:   schoolRepo,
	}
}

// SubmitSIRequest is the payload from SIMAK
type SubmitSIRequest struct {
	SINumber       string          `json:"si_number"`
	IdempotencyKey string          `json:"idempotency_key"`
	BatchID        string          `json:"batch_id"`
	Notes          string          `json:"notes"`
	Items          []SIItemRequest `json:"items"`
}

type SIItemRequest struct {
	NasabahName       string `json:"nasabah_name"`
	NasabahIdentifier string `json:"nasabah_identifier"`
	NasabahType       string `json:"nasabah_type"` // "siswa" or "gtk"
	Amount            int64  `json:"amount"`
}

// SubmitSI processes an SI submission from SIMAK (called via HMAC-protected endpoint)
func (s *SIService) SubmitSI(schoolID uuid.UUID, req *SubmitSIRequest) (*model.SIDocument, error) {
	// Check idempotency — if already exists, return existing document
	existing, err := s.siRepo.FindDocumentByIdempotencyKey(req.IdempotencyKey)
	if err == nil && existing != nil {
		return existing, nil
	}

	// Validate items
	if len(req.Items) == 0 {
		return nil, errors.New("items tidak boleh kosong")
	}
	if len(req.Items) > 200 {
		return nil, errors.New("maksimal 200 item per batch, silakan gunakan chunking")
	}

	// Calculate totals
	var totalAmount int64
	for _, item := range req.Items {
		if item.Amount <= 0 {
			return nil, fmt.Errorf("nominal untuk %s harus positif", item.NasabahName)
		}
		totalAmount += item.Amount
	}

	// Create SI Document within a transaction
	var doc *model.SIDocument

	err = s.siRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		doc = &model.SIDocument{
			SchoolID:       schoolID,
			SINumber:       req.SINumber,
			IdempotencyKey: req.IdempotencyKey,
			BatchID:        req.BatchID,
			Status:         model.SIStatusPending,
			TotalItems:     len(req.Items),
			TotalAmount:    totalAmount,
			Notes:          req.Notes,
		}

		if err := tx.Create(doc).Error; err != nil {
			return fmt.Errorf("gagal membuat dokumen SI: %w", err)
		}

		// Create items
		items := make([]model.SIItem, 0, len(req.Items))
		for _, item := range req.Items {
			items = append(items, model.SIItem{
				SIDocumentID:      doc.ID,
				NasabahName:       item.NasabahName,
				NasabahIdentifier: item.NasabahIdentifier,
				NasabahType:       item.NasabahType,
				Amount:            item.Amount,
				Status:            model.SIItemStatusPending,
			})
		}

		if err := tx.CreateInBatches(items, 100).Error; err != nil {
			return fmt.Errorf("gagal membuat item SI: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	log.Printf("📄 SI diterima: %s dari sekolah %s (%d nasabah, Rp %d)",
		doc.SINumber, schoolID, doc.TotalItems, doc.TotalAmount)

	return doc, nil
}

// VerifySI moves an SI to VERIFIED status
func (s *SIService) VerifySI(siID, adminID uuid.UUID, ip string) error {
	return s.siRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		doc, err := s.siRepo.FindDocumentByIDWithLock(tx, siID)
		if err != nil {
			return errors.New("dokumen SI tidak ditemukan")
		}

		if doc.Status != model.SIStatusPending && doc.Status != model.SIStatusProcessing {
			return fmt.Errorf("dokumen SI tidak bisa diverifikasi dari status %s", doc.Status)
		}

		now := time.Now()
		updates := map[string]interface{}{
			"status":      model.SIStatusVerified,
			"verified_by": adminID,
			"verified_at": now,
		}

		if err := tx.Model(&model.SIDocument{}).Where("id = ?", siID).Updates(updates).Error; err != nil {
			return err
		}

		s.logAudit(&adminID, &siID, "VERIFY_SI", doc.Status, model.SIStatusVerified, ip)
		return nil
	})
}

// ApproveSI moves an SI to APPROVED status
func (s *SIService) ApproveSI(siID, adminID uuid.UUID, ip string) error {
	return s.siRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		doc, err := s.siRepo.FindDocumentByIDWithLock(tx, siID)
		if err != nil {
			return errors.New("dokumen SI tidak ditemukan")
		}

		if doc.Status != model.SIStatusVerified {
			return fmt.Errorf("dokumen SI harus diverifikasi terlebih dahulu (status saat ini: %s)", doc.Status)
		}

		now := time.Now()
		updates := map[string]interface{}{
			"status":      model.SIStatusApproved,
			"approved_by": adminID,
			"approved_at": now,
		}

		if err := tx.Model(&model.SIDocument{}).Where("id = ?", siID).Updates(updates).Error; err != nil {
			return err
		}

		s.logAudit(&adminID, &siID, "APPROVE_SI", doc.Status, model.SIStatusApproved, ip)
		return nil
	})
}

// DisburseSI moves an SI to DISBURSED status and triggers callback
func (s *SIService) DisburseSI(siID, adminID uuid.UUID, ip string) error {
	var doc *model.SIDocument
	err := s.siRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		var txErr error
		doc, txErr = s.siRepo.FindDocumentByIDWithLock(tx, siID)
		if txErr != nil {
			return errors.New("dokumen SI tidak ditemukan")
		}

		if doc.Status != model.SIStatusApproved {
			return fmt.Errorf("dokumen SI harus di-approve terlebih dahulu (status saat ini: %s)", doc.Status)
		}

		now := time.Now()
		updates := map[string]interface{}{
			"status":       model.SIStatusDisbursed,
			"disbursed_at": now,
		}

		if txErr := tx.Model(&model.SIDocument{}).Where("id = ?", siID).Updates(updates).Error; txErr != nil {
			return txErr
		}

		// Update all items to PROCESSED
		for _, item := range doc.Items {
			if txErr := tx.Model(&model.SIItem{}).Where("id = ?", item.ID).Update("status", model.SIItemStatusProcessed).Error; txErr != nil {
				return txErr
			}
		}

		s.logAudit(&adminID, &siID, "DISBURSE_SI", doc.Status, model.SIStatusDisbursed, ip)
		doc.Status = model.SIStatusDisbursed // Update for callback
		return nil
	})

	if err != nil {
		return err
	}

	// Queue callback to SIMAK
	s.queueCallback(doc, "DISBURSED")
	return nil
}

// RejectSI moves an SI to REJECTED status
func (s *SIService) RejectSI(siID, adminID uuid.UUID, reason, ip string) error {
	var doc *model.SIDocument
	err := s.siRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		var txErr error
		doc, txErr = s.siRepo.FindDocumentByIDWithLock(tx, siID)
		if txErr != nil {
			return errors.New("dokumen SI tidak ditemukan")
		}

		if doc.Status == model.SIStatusDisbursed {
			return errors.New("dokumen SI yang sudah dicairkan tidak bisa ditolak")
		}

		updates := map[string]interface{}{
			"status": model.SIStatusRejected,
			"notes":  reason,
		}

		if txErr := tx.Model(&model.SIDocument{}).Where("id = ?", siID).Updates(updates).Error; txErr != nil {
			return txErr
		}

		s.logAudit(&adminID, &siID, "REJECT_SI", doc.Status, model.SIStatusRejected, ip)
		doc.Status = model.SIStatusRejected // Update for callback
		return nil
	})

	if err != nil {
		return err
	}

	// Queue callback to SIMAK
	s.queueCallback(doc, "REJECTED")
	return nil
}

// ListDocuments returns paginated SI documents with optional filters
func (s *SIService) ListDocuments(page, perPage int, status string, schoolID *uuid.UUID) ([]model.SIDocument, int64, error) {
	return s.siRepo.FindAllDocuments(page, perPage, status, schoolID)
}

// GetDocumentByID returns a single SI document with all relations
func (s *SIService) GetDocumentByID(id uuid.UUID) (*model.SIDocument, error) {
	return s.siRepo.FindDocumentByID(id)
}

// GetDashboard returns aggregated dashboard statistics
func (s *SIService) GetDashboard() (*repository.DashboardStats, error) {
	return s.siRepo.GetDashboardStats()
}

// ---- Private helpers ----

func (s *SIService) logAudit(adminID, siDocID *uuid.UUID, action string, oldStatus, newStatus interface{}, ip string) {
	oldData, _ := json.Marshal(map[string]interface{}{"status": oldStatus})
	newData, _ := json.Marshal(map[string]interface{}{"status": newStatus})

	s.auditRepo.Create(&model.AuditLog{
		AdminID:      adminID,
		SIDocumentID: siDocID,
		Action:       action,
		OldData:      string(oldData),
		NewData:      string(newData),
		IPAddress:    ip,
	})
}

func (s *SIService) queueCallback(doc *model.SIDocument, status string) {
	school, err := s.schoolRepo.FindByID(doc.SchoolID)
	if err != nil || school.CallbackURL == "" {
		log.Printf("⚠️ Tidak bisa kirim callback: sekolah %s tidak memiliki callback URL", doc.SchoolID)
		return
	}

	// Otomatisasi URL agar admin hanya perlu isi domain
	baseURL := strings.TrimRight(school.CallbackURL, "/")
	if !strings.HasSuffix(baseURL, "/api/v1/bank-sampah/callback") {
		baseURL += "/api/v1/bank-sampah/callback"
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"si_id":     doc.ID,
		"si_number": doc.SINumber,
		"status":    status,
		"total":     doc.TotalAmount,
		"items":     doc.TotalItems,
		"timestamp": time.Now().Unix(),
	})

	now := time.Now()
	s.callbackRepo.Create(&model.CallbackQueue{
		SIDocumentID: doc.ID,
		CallbackURL:  baseURL,
		Payload:      string(payload),
		Status:       model.CallbackStatusPending,
		NextRetryAt:  &now,
	})

	log.Printf("📤 Callback queued untuk SI %s → %s", doc.SINumber, school.CallbackURL)
}
