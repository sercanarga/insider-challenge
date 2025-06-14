package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"insider-challenge/internal/repository"
	"insider-challenge/pkg/config"
	domain "insider-challenge/pkg/domain"
	"insider-challenge/pkg/errors"
)

const (
	defaultMessageBatchSize = 2
	defaultTickerInterval   = 2 * time.Minute
	defaultHTTPTimeout      = 10 * time.Second
	defaultRequestTimeout   = 5 * time.Second
)

// Service handles the business logic for message process
type Service struct {
	repo        repository.Repository
	cfg         *config.Config
	stopChan    chan struct{}
	doneChan    chan struct{}
	isRunning   bool
	runningLock sync.Mutex
}

// New creates a new service instance
func New(repo repository.Repository, cfg *config.Config) *Service {
	return &Service{
		repo:     repo,
		cfg:      cfg,
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
	}
}

// StartMessageSender starts the message sender service, periodically sends unsent messages
// @note: it runs in a goroutine and can be stopped using StopMessageSender.
func (s *Service) StartMessageSender() {
	s.runningLock.Lock()
	if s.isRunning {
		s.runningLock.Unlock()
		return
	}
	s.isRunning = true
	s.stopChan = make(chan struct{})
	s.doneChan = make(chan struct{})
	s.runningLock.Unlock()

	go func() {
		ticker := time.NewTicker(defaultTickerInterval)
		defer ticker.Stop()
		defer close(s.doneChan)

		for {
			select {
			case <-ticker.C:
				if err := s.sendMessages(); err != nil {
					log.Printf("Failed to send messages: %v", err)
				}
			case <-s.stopChan:
				s.runningLock.Lock()
				s.isRunning = false
				s.runningLock.Unlock()
				return
			}
		}
	}()
}

// StopMessageSender stops the message sender service gracefully
func (s *Service) StopMessageSender() {
	s.runningLock.Lock()
	if !s.isRunning {
		s.runningLock.Unlock()
		return
	}

	if s.stopChan != nil {
		close(s.stopChan)
		s.stopChan = nil
	}
	s.isRunning = false
	s.runningLock.Unlock()

	// Wait for the service to fully stop
	<-s.doneChan
}

// sendMessages retrieves and sends unsent messages in batches
func (s *Service) sendMessages() error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultHTTPTimeout)
	defer cancel()

	messages, err := s.repo.GetUnsentMessages(ctx, defaultMessageBatchSize)
	if err != nil {
		return errors.Wrap(err, "get unsent messages")
	}

	for _, msg := range messages {
		if err := s.sendMessage(ctx, msg); err != nil {
			log.Printf("Failed to send message %s: %v", msg.ID, err)
			continue
		}

		if err := s.repo.MarkMessageAsSent(ctx, msg.ID.String()); err != nil {
			log.Printf("Failed to mark message %s as sent: %v", msg.ID, err)
		}
	}

	return nil
}

// sendMessage sends a single message to the configured webhook uri
func (s *Service) sendMessage(ctx context.Context, msg domain.Message) error {
	reqCtx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()

	payload := map[string]string{
		"to":      msg.To,
		"content": msg.Content,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "marshal payload")
	}

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, s.cfg.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return errors.Wrap(err, "create request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-ins-auth-key", s.cfg.WebhookAuthKey)

	client := &http.Client{
		Timeout: defaultHTTPTimeout,
		Transport: &http.Transport{
			ResponseHeaderTimeout: defaultRequestTimeout,
			ExpectContinueTimeout: defaultRequestTimeout,
			IdleConnTimeout:       defaultRequestTimeout,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		if reqCtx.Err() == context.DeadlineExceeded {
			return errors.Wrap(errors.ErrWebhookFailed, "request timeout exceeded")
		}
		return errors.Wrap(err, "send request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return errors.Wrap(errors.ErrWebhookFailed, fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
	}

	// Parse the response to get the message, messageId
	var response struct {
		Message   string `json:"message"`
		MessageID string `json:"messageId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return errors.Wrap(err, "decode response")
	}

	// Cache the messageId in redis after sending
	if err := config.CacheMessageID(ctx, msg.ID.String(), response.MessageID); err != nil {
		log.Printf("Failed to cache message ID %s: %v", msg.ID, err)
	}

	return nil
}

// GetSentMessages retrieves all sent messages from the repository
func (s *Service) GetSentMessages(page, pageSize int) ([]domain.Message, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultHTTPTimeout)
	defer cancel()

	messages, total, err := s.repo.GetSentMessages(ctx, page, pageSize)
	if err != nil {
		return nil, 0, errors.Wrap(err, "get sent messages")
	}
	return messages, total, nil
}

// IsRunning returns whether the message sender is currently running
func (s *Service) IsRunning() bool {
	s.runningLock.Lock()
	defer s.runningLock.Unlock()
	return s.isRunning
}
