package service

import (
	"context"
	"time"

	"insider-challenge/internal/repository"
	"insider-challenge/pkg/config"
	domain "insider-challenge/pkg/domain"
	"insider-challenge/pkg/errors"
)

// Service handles the business logic for message process
type Service struct {
	repo          repository.Repository
	cfg           *config.Config
	messageSender *MessageSender
	httpTimeout   time.Duration
}

// New creates a new service instance
func New(repo repository.Repository, cfg *config.Config) *Service {
	messageSender := NewMessageSender(repo, cfg)
	return &Service{
		repo:          repo,
		cfg:           cfg,
		messageSender: messageSender,
		httpTimeout:   10 * time.Second,
	}
}

// StartMessageSender starts the message sender service
func (s *Service) StartMessageSender() {
	s.messageSender.Start()
}

// StopMessageSender stops the message sender service gracefully
func (s *Service) StopMessageSender() {
	s.messageSender.Stop()
}

// GetSentMessages retrieves all sent messages from the repository
func (s *Service) GetSentMessages(page, pageSize int) ([]domain.Message, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.httpTimeout)
	defer cancel()

	messages, total, err := s.repo.GetSentMessages(ctx, page, pageSize)
	if err != nil {
		return nil, 0, errors.Wrap(err, "get sent messages")
	}
	return messages, total, nil
}

// IsRunning returns whether the message sender is currently running
func (s *Service) IsRunning() bool {
	return s.messageSender.IsRunning()
}
