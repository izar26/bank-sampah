package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SIStatus represents the state of a Surat Instruksi
type SIStatus string

const (
	SIStatusPending    SIStatus = "PENDING"
	SIStatusProcessing SIStatus = "PROCESSING"
	SIStatusVerified   SIStatus = "VERIFIED"
	SIStatusApproved   SIStatus = "APPROVED"
	SIStatusDisbursed  SIStatus = "DISBURSED"
	SIStatusRejected   SIStatus = "REJECTED"
)

// SIDocument represents a Surat Instruksi (instruction letter) from a school
type SIDocument struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID       uuid.UUID      `gorm:"type:uuid;not null;index" json:"school_id"`
	SINumber       string         `gorm:"uniqueIndex;size:100;not null;column:si_number" json:"si_number"`
	IdempotencyKey string         `gorm:"uniqueIndex;size:100;not null" json:"idempotency_key"`
	BatchID        string         `gorm:"size:100;index" json:"batch_id"`
	Status         SIStatus       `gorm:"size:20;not null;default:PENDING;index" json:"status"`
	TotalItems     int            `gorm:"not null;default:0" json:"total_items"`
	TotalAmount    int64          `gorm:"not null;default:0" json:"total_amount"`
	Notes          string         `gorm:"type:text" json:"notes"`
	VerifiedBy     *uuid.UUID     `gorm:"type:uuid" json:"verified_by"`
	ApprovedBy     *uuid.UUID     `gorm:"type:uuid" json:"approved_by"`
	VerifiedAt     *time.Time     `json:"verified_at"`
	ApprovedAt     *time.Time     `json:"approved_at"`
	DisbursedAt    *time.Time     `json:"disbursed_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	School       School          `gorm:"foreignKey:SchoolID" json:"school,omitempty"`
	Items        []SIItem        `gorm:"foreignKey:SIDocumentID" json:"items,omitempty"`
	AuditLogs    []AuditLog      `gorm:"foreignKey:SIDocumentID" json:"audit_logs,omitempty"`
	Verifier     *Admin          `gorm:"foreignKey:VerifiedBy" json:"verifier,omitempty"`
	Approver     *Admin          `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
}

func (s *SIDocument) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (SIDocument) TableName() string {
	return "si_documents"
}
