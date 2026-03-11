package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SMSCode represents a temporary sms verification code.
type SMSCode struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey"`
	Phone     string    `gorm:"size:32;index;not null"`
	Purpose   string    `gorm:"size:32;index;not null"`
	CodeHash  string    `gorm:"size:255;not null"`
	ExpiresAt time.Time `gorm:"index;not null"`
	Used      bool      `gorm:"default:false;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// BeforeCreate assigns UUIDs automatically.
func (s *SMSCode) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
