package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditLog records every significant action for compliance and traceability
type AuditLog struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	AdminID      *uuid.UUID `gorm:"type:uuid;index" json:"admin_id"`
	SIDocumentID *uuid.UUID `gorm:"type:uuid;index" json:"si_document_id"`
	Action       string     `gorm:"size:100;not null;index" json:"action"`
	OldData      string     `gorm:"type:jsonb" json:"old_data,omitempty"`
	NewData      string     `gorm:"type:jsonb" json:"new_data,omitempty"`
	IPAddress    string     `gorm:"size:45" json:"ip_address"`
	UserAgent    string     `gorm:"size:500" json:"user_agent"`
	CreatedAt    time.Time  `gorm:"index" json:"created_at"`

	// Relations
	Admin      *Admin      `gorm:"foreignKey:AdminID" json:"admin,omitempty"`
	SIDocument *SIDocument `gorm:"foreignKey:SIDocumentID" json:"si_document,omitempty"`
}

func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
