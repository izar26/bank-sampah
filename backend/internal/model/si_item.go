package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SIItemStatus represents the processing state of individual nasabah item
type SIItemStatus string

const (
	SIItemStatusPending   SIItemStatus = "PENDING"
	SIItemStatusProcessed SIItemStatus = "PROCESSED"
	SIItemStatusFailed    SIItemStatus = "FAILED"
)

// SIItem represents a single nasabah entry within an SI document
type SIItem struct {
	ID                 uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	SIDocumentID       uuid.UUID    `gorm:"type:uuid;not null;index" json:"si_document_id"`
	NasabahName        string       `gorm:"size:200;not null" json:"nasabah_name"`
	NasabahIdentifier  string       `gorm:"size:100;not null" json:"nasabah_identifier"`
	NasabahType        string       `gorm:"size:10;not null" json:"nasabah_type"` // "siswa" or "gtk"
	Amount             int64        `gorm:"not null" json:"amount"`
	Status             SIItemStatus `gorm:"size:20;not null;default:PENDING;index" json:"status"`
	FailureReason      string       `gorm:"type:text" json:"failure_reason,omitempty"`
	ProcessedAt        *time.Time   `json:"processed_at"`
	CreatedAt          time.Time    `json:"created_at"`
	UpdatedAt          time.Time    `json:"updated_at"`

	// Relations
	SIDocument SIDocument `gorm:"foreignKey:SIDocumentID" json:"si_document,omitempty"`
}

func (s *SIItem) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (SIItem) TableName() string {
	return "si_items"
}
