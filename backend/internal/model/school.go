package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// School represents a registered school (SIMAK tenant) with API credentials
type School struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string         `gorm:"size:200;not null" json:"name"`
	Code        string         `gorm:"uniqueIndex;size:50;not null" json:"code"`
	APIKey      string         `gorm:"uniqueIndex;size:64;not null;column:api_key" json:"api_key"`
	APISecret   string         `gorm:"size:128;not null;column:api_secret" json:"api_secret"`
	CallbackURL string         `gorm:"size:500;column:callback_url" json:"callback_url"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	SIDocuments []SIDocument `gorm:"foreignKey:SchoolID" json:"si_documents,omitempty"`
}

func (s *School) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (School) TableName() string {
	return "schools"
}
