package service

import (
	"log"
	"time"
)

// WorkerPool manages background goroutines for batch processing
type WorkerPool struct {
	callbackSvc *CallbackService
	poolSize    int
	stopCh      chan struct{}
}

func NewWorkerPool(callbackSvc *CallbackService, poolSize int) *WorkerPool {
	return &WorkerPool{
		callbackSvc: callbackSvc,
		poolSize:    poolSize,
		stopCh:      make(chan struct{}),
	}
}

// Start launches background workers
func (wp *WorkerPool) Start() {
	log.Printf("🚀 Starting worker pool with %d workers", wp.poolSize)

	// Worker 1: Callback dispatcher (runs every 5 seconds)
	go wp.callbackWorker()

	// Worker 2: Nonce cleanup (runs every 10 minutes)
	go wp.nonceCleanupWorker()
}

// Stop gracefully shuts down all workers
func (wp *WorkerPool) Stop() {
	log.Println("🛑 Stopping worker pool...")
	close(wp.stopCh)
}

func (wp *WorkerPool) callbackWorker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-wp.stopCh:
			log.Println("⏹️ Callback worker stopped")
			return
		case <-ticker.C:
			wp.callbackSvc.ProcessPendingCallbacks()
		}
	}
}

func (wp *WorkerPool) nonceCleanupWorker() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-wp.stopCh:
			log.Println("⏹️ Nonce cleanup worker stopped")
			return
		case <-ticker.C:
			wp.callbackSvc.CleanupNonces()
		}
	}
}
