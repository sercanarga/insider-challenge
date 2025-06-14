package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Message structure
type Message struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	To        string         `gorm:"not null" json:"to"`
	Content   string         `gorm:"not null;size:150" json:"content"` // Maximum 150 character (character limit is required for message content)
	IsSent    bool           `gorm:"default:false;index" json:"is_sent"`
	SentAt    *time.Time     `gorm:"index" json:"sent_at"`
	CreatedAt time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
