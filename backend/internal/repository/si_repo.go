package repository

import (
	"bank-sampah-backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SIRepository struct {
	db *gorm.DB
}

func NewSIRepository(db *gorm.DB) *SIRepository {
	return &SIRepository{db: db}
}

// ---- SI Document ----

func (r *SIRepository) FindAllDocuments(page, perPage int, status string, schoolID *uuid.UUID) ([]model.SIDocument, int64, error) {
	var documents []model.SIDocument
	var total int64

	query := r.db.Model(&model.SIDocument{}).Preload("School")

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if schoolID != nil {
		query = query.Where("school_id = ?", *schoolID)
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&documents).Error

	return documents, total, err
}

func (r *SIRepository) FindDocumentByID(id uuid.UUID) (*model.SIDocument, error) {
	var doc model.SIDocument
	err := r.db.Preload("School").Preload("Items").Preload("Verifier").Preload("Approver").
		First(&doc, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *SIRepository) FindDocumentByIDWithLock(tx *gorm.DB, id uuid.UUID) (*model.SIDocument, error) {
	var doc model.SIDocument
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("School").Preload("Items").Preload("Verifier").Preload("Approver").
		First(&doc, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *SIRepository) FindDocumentByIdempotencyKey(key string) (*model.SIDocument, error) {
	var doc model.SIDocument
	err := r.db.Where("idempotency_key = ?", key).First(&doc).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *SIRepository) CreateDocument(doc *model.SIDocument) error {
	return r.db.Create(doc).Error
}

func (r *SIRepository) UpdateDocumentStatus(id uuid.UUID, status model.SIStatus, updates map[string]interface{}) error {
	updates["status"] = status
	return r.db.Model(&model.SIDocument{}).Where("id = ?", id).Updates(updates).Error
}

// ---- SI Items ----

func (r *SIRepository) CreateItems(items []model.SIItem) error {
	return r.db.CreateInBatches(items, 100).Error // Batch insert 100 at a time
}

func (r *SIRepository) FindPendingItems(limit int) ([]model.SIItem, error) {
	var items []model.SIItem
	err := r.db.Where("status = ?", model.SIItemStatusPending).
		Limit(limit).
		Find(&items).Error
	return items, err
}

func (r *SIRepository) UpdateItemStatus(id uuid.UUID, status model.SIItemStatus, reason string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if reason != "" {
		updates["failure_reason"] = reason
	}
	if status == model.SIItemStatusProcessed || status == model.SIItemStatusFailed {
		updates["processed_at"] = gorm.Expr("NOW()")
	}
	return r.db.Model(&model.SIItem{}).Where("id = ?", id).Updates(updates).Error
}

func (r *SIRepository) CountItemsByDocumentAndStatus(docID uuid.UUID) (map[string]int64, error) {
	type Result struct {
		Status string
		Count  int64
	}

	var results []Result
	err := r.db.Model(&model.SIItem{}).
		Select("status, COUNT(*) as count").
		Where("si_document_id = ?", docID).
		Group("status").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, r := range results {
		counts[r.Status] = r.Count
	}
	return counts, nil
}

// ---- Dashboard Stats ----

type DashboardStats struct {
	TotalSchools      int64 `json:"total_schools"`
	TotalSI           int64 `json:"total_si"`
	PendingSI         int64 `json:"pending_si"`
	ProcessingSI      int64 `json:"processing_si"`
	VerifiedSI        int64 `json:"verified_si"`
	ApprovedSI        int64 `json:"approved_si"`
	DisbursedSI       int64 `json:"disbursed_si"`
	RejectedSI        int64 `json:"rejected_si"`
	TotalDisbursed    int64 `json:"total_disbursed"`
	TotalNasabah      int64 `json:"total_nasabah"`
}

func (r *SIRepository) GetDashboardStats() (*DashboardStats, error) {
	stats := &DashboardStats{}

	r.db.Model(&model.School{}).Where("is_active = true").Count(&stats.TotalSchools)
	r.db.Model(&model.SIDocument{}).Count(&stats.TotalSI)
	r.db.Model(&model.SIDocument{}).Where("status = ?", model.SIStatusPending).Count(&stats.PendingSI)
	r.db.Model(&model.SIDocument{}).Where("status = ?", model.SIStatusProcessing).Count(&stats.ProcessingSI)
	r.db.Model(&model.SIDocument{}).Where("status = ?", model.SIStatusVerified).Count(&stats.VerifiedSI)
	r.db.Model(&model.SIDocument{}).Where("status = ?", model.SIStatusApproved).Count(&stats.ApprovedSI)
	r.db.Model(&model.SIDocument{}).Where("status = ?", model.SIStatusDisbursed).Count(&stats.DisbursedSI)
	r.db.Model(&model.SIDocument{}).Where("status = ?", model.SIStatusRejected).Count(&stats.RejectedSI)

	// Total amount disbursed
	r.db.Model(&model.SIDocument{}).
		Where("status = ?", model.SIStatusDisbursed).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&stats.TotalDisbursed)

	// Total unique nasabah items
	r.db.Model(&model.SIItem{}).Count(&stats.TotalNasabah)

	return stats, nil
}

// GetDB exposes the database for transaction use
func (r *SIRepository) GetDB() *gorm.DB {
	return r.db
}
