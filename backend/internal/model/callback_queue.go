package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CallbackStatus represents the delivery state of a callback
type CallbackStatus string

const (
	CallbackStatusPending    CallbackStatus = "PENDING"
	CallbackStatusSuccess    CallbackStatus = "SUCCESS"
	CallbackStatusFailed     CallbackStatus = "FAILED"
	CallbackStatusDeadLetter CallbackStatus = "DEAD_LETTER"
)

// CallbackQueue stores outbound callbacks to SIMAK with retry tracking
type CallbackQueue struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	SIDocumentID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"si_document_id"`
	CallbackURL     string         `gorm:"size:500;not null" json:"callback_url"`
	Payload         string         `gorm:"type:jsonb;not null" json:"payload"`
	RetryCount      int            `gorm:"not null;default:0" json:"retry_count"`
	Status          CallbackStatus `gorm:"size:20;not null;default:PENDING;index" json:"status"`
	NextRetryAt     *time.Time     `gorm:"index" json:"next_retry_at"`
	LastAttemptedAt *time.Time     `json:"last_attempted_at"`
	LastError       string         `gorm:"type:text" json:"last_error,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`

	// Relations
	SIDocument SIDocument `gorm:"foreignKey:SIDocumentID" json:"si_document,omitempty"`
}

func (c *CallbackQueue) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (CallbackQueue) TableName() string {
	return "callback_queue"
}
