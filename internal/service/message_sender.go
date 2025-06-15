package service

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"insider-challenge/internal/repository"
	"insider-challenge/pkg/config"
	domain "insider-challenge/pkg/domain"
	"insider-challenge/pkg/errors"
)

// MessageSender handles the message sending
type MessageSender struct {
	repo        repository.Repository
	cfg         *config.Config
	httpClient  *HTTPClient
	stopChan    chan struct{}
	doneChan    chan struct{}
	isRunning   bool
	runningLock sync.Mutex

	// DefaultMessageBatchSize default number of messages batch
	messageBatchSize int

	// DefaultTickerInterval default interval sender ticker
	tickerInterval time.Duration

	// DefaultHTTPTimeout default timeout http operations
	httpTimeout time.Duration

	// DefaultRequestTimeout default timeout for requests
	requestTimeout time.Duration
}

// NewMessageSender creates a new message sender instance
func NewMessageSender(repo repository.Repository, cfg *config.Config) *MessageSender {
	return &MessageSender{
		repo:             repo,
		cfg:              cfg,
		httpClient:       NewHTTPClient(cfg),
		stopChan:         make(chan struct{}),
		doneChan:         make(chan struct{}),
		messageBatchSize: 2,
		tickerInterval:   2 * time.Minute,
		httpTimeout:      10 * time.Second,
		requestTimeout:   5 * time.Second,
	}
}

// Start starts the message sender service
func (ms *MessageSender) Start() {
	ms.runningLock.Lock()
	if ms.isRunning {
		ms.runningLock.Unlock()
		return
	}
	ms.isRunning = true
	ms.stopChan = make(chan struct{})
	ms.doneChan = make(chan struct{})
	ms.runningLock.Unlock()

	go func() {
		ticker := time.NewTicker(ms.tickerInterval)
		defer ticker.Stop()
		defer close(ms.doneChan)

		for {
			select {
			case <-ticker.C:
				if err := ms.sendMessages(); err != nil {
					log.Printf("Failed to send messages: %v", err)
				}
			case <-ms.stopChan:
				ms.runningLock.Lock()
				ms.isRunning = false
				ms.runningLock.Unlock()
				return
			}
		}
	}()
}

// Stop stops the message sender service gracefully
func (ms *MessageSender) Stop() {
	ms.runningLock.Lock()
	if !ms.isRunning {
		ms.runningLock.Unlock()
		return
	}

	if ms.stopChan != nil {
		close(ms.stopChan)
		ms.stopChan = nil
	}
	ms.isRunning = false
	ms.runningLock.Unlock()

	<-ms.doneChan
}

// IsRunning returns whether the message sender is currently running
func (ms *MessageSender) IsRunning() bool {
	ms.runningLock.Lock()
	defer ms.runningLock.Unlock()
	return ms.isRunning
}

// sendMessages retrieves and sends unsent messages in batches
func (ms *MessageSender) sendMessages() error {
	ctx, cancel := context.WithTimeout(context.Background(), ms.httpTimeout)
	defer cancel()

	messages, err := ms.repo.GetUnsentMessages(ctx, ms.messageBatchSize)
	if err != nil {
		return errors.Wrap(err, "get unsent messages")
	}

	for _, msg := range messages {
		if err := ms.sendMessage(ctx, msg); err != nil {
			log.Printf("Failed to send message %s: %v", msg.ID, err)
			continue
		}

		if err := ms.repo.MarkMessageAsSent(ctx, msg.ID.String()); err != nil {
			log.Printf("Failed to mark message %s as sent: %v", msg.ID, err)
		}
	}

	return nil
}

// sendMessage sends a single message to the configured webhook uri
func (ms *MessageSender) sendMessage(ctx context.Context, msg domain.Message) error {
	reqCtx, cancel := context.WithTimeout(ctx, ms.requestTimeout)
	defer cancel()

	payload := map[string]string{
		"to":      msg.To,
		"content": msg.Content,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "marshal payload")
	}

	response, err := ms.httpClient.SendRequest(reqCtx, jsonData)
	if err != nil {
		return err
	}

	if err := config.CacheMessageID(ctx, msg.ID.String(), response.MessageID); err != nil {
		log.Printf("Failed to cache message ID %s: %v", msg.ID, err)
	}

	return nil
}
