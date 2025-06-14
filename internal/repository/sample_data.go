package repository

import (
	"insider-challenge/pkg/domain"
	"time"

	"gorm.io/gorm"
)

// InitSampleData creates sample messages if the db is empty
func InitSampleData(db *gorm.DB) error {
	var count int64
	if err := db.Model(&domain.Message{}).Count(&count).Error; err != nil {
		return err
	}

	if count != 0 {
		return nil
	}

	sampleMessages := []domain.Message{
		{
			To:        "+905071773757",
			Content:   "Merhaba! Bu bir örnek mesajdır.",
			CreatedAt: time.Now(),
		},
		{
			To:        "+90555255555",
			Content:   "İkinci örnek mesaj",
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			To:        "+905071773525",
			Content:   "Üçüncü örnek mesaj",
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
	}

	if err := db.Create(&sampleMessages).Error; err != nil {
		return err
	}

	return nil
}
