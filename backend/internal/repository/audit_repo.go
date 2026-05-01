package repository

import (
	"bank-sampah-backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(log *model.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *AuditRepository) FindAll(page, perPage int, adminID *uuid.UUID) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64

	query := r.db.Model(&model.AuditLog{}).Preload("Admin")

	if adminID != nil {
		query = query.Where("admin_id = ?", *adminID)
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&logs).Error

	return logs, total, err
}

func (r *AuditRepository) FindBySIDocument(siDocID uuid.UUID) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	err := r.db.Preload("Admin").
		Where("si_document_id = ?", siDocID).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}
