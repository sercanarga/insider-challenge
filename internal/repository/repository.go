package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"insider-challenge/pkg/domain"
	"insider-challenge/pkg/errors"
)

// Repository defines the interface
type Repository interface {
	GetUnsentMessages(ctx context.Context, limit int) ([]domain.Message, error)
	MarkMessageAsSent(ctx context.Context, messageID string) error
	GetSentMessages(ctx context.Context, page, pageSize int) ([]domain.Message, int64, error)
	CreateMessage(ctx context.Context, message *domain.Message) error
	GetMessageByID(ctx context.Context, messageID string) (*domain.Message, error)
}

// repository implements the repository interface
type repository struct {
	db *gorm.DB
}

// New create new repository instance
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

// GetUnsentMessages retrieves unsent messages from db
func (r *repository) GetUnsentMessages(ctx context.Context, limit int) ([]domain.Message, error) {
	var messages []domain.Message
	err := r.db.WithContext(ctx).
		Where("is_sent = ? AND deleted_at IS NULL", false).
		Order("created_at ASC").
		Limit(limit).
		Find(&messages).Error
	if err != nil {
		return nil, errors.Wrap(err, "get unsent messages")
	}
	return messages, nil
}

// MarkMessageAsSent marks message as sent in the database
func (r *repository) MarkMessageAsSent(ctx context.Context, messageID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var message domain.Message
		if err := tx.Where("id = ? AND deleted_at IS NULL", messageID).First(&message).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.Wrap(errors.ErrMessageNotFound, fmt.Sprintf("message ID: %s", messageID))
			}
			return errors.Wrap(err, "find message")
		}

		now := time.Now()
		updates := map[string]interface{}{
			"is_sent": true,
			"sent_at": now,
		}

		if err := tx.Model(&message).Updates(updates).Error; err != nil {
			return errors.Wrap(err, "update message")
		}

		return nil
	})
}

// GetSentMessages retrieves all sent messages from the db
func (r *repository) GetSentMessages(ctx context.Context, page, pageSize int) ([]domain.Message, int64, error) {
	var messages []domain.Message
	var total int64

	offset := (page - 1) * pageSize

	// Get total count
	err := r.db.WithContext(ctx).
		Model(&domain.Message{}).
		Where("is_sent = ? AND deleted_at IS NULL", true).
		Count(&total).Error
	if err != nil {
		return nil, 0, errors.Wrap(err, "count sent messages")
	}

	err = r.db.WithContext(ctx).
		Where("is_sent = ? AND deleted_at IS NULL", true).
		Order("sent_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&messages).Error
	if err != nil {
		return nil, 0, errors.Wrap(err, "get sent messages")
	}

	return messages, total, nil
}

// CreateMessage create new message in the db
func (r *repository) CreateMessage(ctx context.Context, message *domain.Message) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(message).Error; err != nil {
			return errors.Wrap(err, "create message")
		}
		return nil
	})
}

// GetMessageByID retrieves a message by id
// @info: unused, sample helper function
func (r *repository) GetMessageByID(ctx context.Context, messageID string) (*domain.Message, error) {
	var message domain.Message
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", messageID).
		First(&message).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.Wrap(errors.ErrMessageNotFound, fmt.Sprintf("message ID: %s", messageID))
		}
		return nil, errors.Wrap(err, "get message by ID")
	}
	return &message, nil
}
