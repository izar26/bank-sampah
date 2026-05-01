package repository

import (
	"time"

	"bank-sampah-backend/internal/model"

	"gorm.io/gorm"
)

type CallbackRepository struct {
	db *gorm.DB
}

func NewCallbackRepository(db *gorm.DB) *CallbackRepository {
	return &CallbackRepository{db: db}
}

func (r *CallbackRepository) Create(cb *model.CallbackQueue) error {
	return r.db.Create(cb).Error
}

func (r *CallbackRepository) FindPendingCallbacks(limit int) ([]model.CallbackQueue, error) {
	var callbacks []model.CallbackQueue
	now := time.Now()
	err := r.db.Where("status = ? AND (next_retry_at IS NULL OR next_retry_at <= ?)",
		model.CallbackStatusPending, now).
		Order("created_at ASC").
		Limit(limit).
		Find(&callbacks).Error
	return callbacks, err
}

func (r *CallbackRepository) MarkSuccess(cb *model.CallbackQueue) error {
	now := time.Now()
	return r.db.Model(cb).Updates(map[string]interface{}{
		"status":           model.CallbackStatusSuccess,
		"last_attempted_at": now,
	}).Error
}

func (r *CallbackRepository) MarkFailed(cb *model.CallbackQueue, lastError string, maxRetries int) error {
	now := time.Now()
	cb.RetryCount++
	cb.LastAttemptedAt = &now
	cb.LastError = lastError

	if cb.RetryCount >= maxRetries {
		cb.Status = model.CallbackStatusDeadLetter
	} else {
		cb.Status = model.CallbackStatusPending
		// Exponential backoff: 2^retryCount seconds
		backoff := time.Duration(1<<uint(cb.RetryCount)) * time.Second
		nextRetry := now.Add(backoff)
		cb.NextRetryAt = &nextRetry
	}

	return r.db.Save(cb).Error
}

func (r *CallbackRepository) FindDeadLetters(page, perPage int) ([]model.CallbackQueue, int64, error) {
	var callbacks []model.CallbackQueue
	var total int64

	query := r.db.Model(&model.CallbackQueue{}).
		Where("status = ?", model.CallbackStatusDeadLetter)

	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&callbacks).Error

	return callbacks, total, err
}

// CleanExpiredNonces removes old nonces that are past their expiry
func (r *CallbackRepository) CleanExpiredNonces() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&model.UsedNonce{}).Error
}

// SaveNonce stores a nonce to prevent replay attacks
func (r *CallbackRepository) SaveNonce(nonce string, ttl time.Duration) error {
	return r.db.Create(&model.UsedNonce{
		Nonce:     nonce,
		ExpiresAt: time.Now().Add(ttl),
	}).Error
}

// NonceExists checks if a nonce has already been used
func (r *CallbackRepository) NonceExists(nonce string) (bool, error) {
	var count int64
	err := r.db.Model(&model.UsedNonce{}).
		Where("nonce = ? AND expires_at > ?", nonce, time.Now()).
		Count(&count).Error
	return count > 0, err
}
