package repository

import (
	"crypto/rand"
	"encoding/hex"

	"bank-sampah-backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SchoolRepository struct {
	db *gorm.DB
}

func NewSchoolRepository(db *gorm.DB) *SchoolRepository {
	return &SchoolRepository{db: db}
}

func (r *SchoolRepository) FindAll() ([]model.School, error) {
	var schools []model.School
	err := r.db.Order("created_at DESC").Find(&schools).Error
	return schools, err
}

func (r *SchoolRepository) FindByID(id uuid.UUID) (*model.School, error) {
	var school model.School
	err := r.db.First(&school, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &school, nil
}

func (r *SchoolRepository) FindByAPIKey(apiKey string) (*model.School, error) {
	var school model.School
	err := r.db.Where("api_key = ? AND is_active = true", apiKey).First(&school).Error
	if err != nil {
		return nil, err
	}
	return &school, nil
}

func (r *SchoolRepository) Create(school *model.School) error {
	// Generate secure API key and secret
	apiKey, err := generateSecureToken(32)
	if err != nil {
		return err
	}
	apiSecret, err := generateSecureToken(64)
	if err != nil {
		return err
	}

	school.APIKey = apiKey
	school.APISecret = apiSecret

	return r.db.Create(school).Error
}

func (r *SchoolRepository) Update(school *model.School) error {
	return r.db.Save(school).Error
}

func (r *SchoolRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.School{}, "id = ?", id).Error
}

// RegenerateCredentials creates new API key and secret for a school
func (r *SchoolRepository) RegenerateCredentials(id uuid.UUID) (*model.School, error) {
	school, err := r.FindByID(id)
	if err != nil {
		return nil, err
	}

	apiKey, err := generateSecureToken(32)
	if err != nil {
		return nil, err
	}
	apiSecret, err := generateSecureToken(64)
	if err != nil {
		return nil, err
	}

	school.APIKey = apiKey
	school.APISecret = apiSecret

	if err := r.db.Save(school).Error; err != nil {
		return nil, err
	}

	return school, nil
}

func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
