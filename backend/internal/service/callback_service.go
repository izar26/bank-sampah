package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"bank-sampah-backend/internal/model"
	"bank-sampah-backend/internal/repository"
)

type CallbackService struct {
	callbackRepo *repository.CallbackRepository
	maxRetries   int
	httpClient   *http.Client
}

func NewCallbackService(callbackRepo *repository.CallbackRepository, maxRetries int) *CallbackService {
	return &CallbackService{
		callbackRepo: callbackRepo,
		maxRetries:   maxRetries,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ProcessPendingCallbacks fetches and dispatches pending callbacks
func (s *CallbackService) ProcessPendingCallbacks() {
	callbacks, err := s.callbackRepo.FindPendingCallbacks(20)
	if err != nil {
		log.Printf("❌ Error fetching pending callbacks: %v", err)
		return
	}

	for _, cb := range callbacks {
		s.dispatchCallback(&cb)
	}
}

func (s *CallbackService) dispatchCallback(cb *model.CallbackQueue) {
	log.Printf("📤 Dispatching callback %s to %s (attempt #%d)", cb.ID, cb.CallbackURL, cb.RetryCount+1)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cb.CallbackURL, bytes.NewBufferString(cb.Payload))
	if err != nil {
		s.handleFailure(cb, err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Callback-ID", cb.ID.String())
	req.Header.Set("X-SI-Document-ID", cb.SIDocumentID.String())

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.handleFailure(cb, err.Error())
		return
	}
	defer resp.Body.Close()

	// Read response body for logging
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		s.callbackRepo.MarkSuccess(cb)
		log.Printf("✅ Callback %s berhasil (status: %d)", cb.ID, resp.StatusCode)
	} else {
		errMsg := string(body)
		if errMsg == "" {
			errMsg = resp.Status
		}
		s.handleFailure(cb, errMsg)
	}
}

func (s *CallbackService) handleFailure(cb *model.CallbackQueue, errMsg string) {
	log.Printf("⚠️ Callback %s gagal: %s", cb.ID, errMsg)
	s.callbackRepo.MarkFailed(cb, errMsg, s.maxRetries)

	if cb.RetryCount >= s.maxRetries {
		log.Printf("💀 Callback %s masuk dead letter queue setelah %d percobaan", cb.ID, s.maxRetries)
	}
}

// CleanupNonces removes expired nonces from the database
func (s *CallbackService) CleanupNonces() {
	if err := s.callbackRepo.CleanExpiredNonces(); err != nil {
		log.Printf("❌ Error cleaning expired nonces: %v", err)
	}
}

// CallbackStats returns stats about callback processing
type CallbackStats struct {
	Pending    int64 `json:"pending"`
	Success    int64 `json:"success"`
	Failed     int64 `json:"failed"`
	DeadLetter int64 `json:"dead_letter"`
}

func (s *CallbackService) GetStats() (*CallbackStats, error) {
	// This is a simple implementation. In production, use a dedicated query.
	pending, _ := s.callbackRepo.FindPendingCallbacks(0)
	_ = pending
	
	// For now, return empty stats — will be enhanced later
	return &CallbackStats{}, nil
}

// SerializePayload helper for creating callback payloads
func SerializePayload(data interface{}) string {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}
