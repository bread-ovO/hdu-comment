package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EmailVerification represents an email verification token
type EmailVerification struct {
	ID        uuid.UUID  `gorm:"type:char(36);primaryKey" json:"id"`
	UserID    *uuid.UUID `gorm:"type:char(36);index" json:"user_id"`
	Email     string     `gorm:"size:255;not null;index" json:"email"`
	Token     string     `gorm:"size:255;not null;uniqueIndex" json:"-"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	Used      bool       `gorm:"default:false" json:"used"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// BeforeCreate hook to set UUIDs automatically
func (e *EmailVerification) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

// IsExpired checks if the verification token has expired
func (e *EmailVerification) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// IsValid checks if the verification token is valid (not used and not expired)
func (e *EmailVerification) IsValid() bool {
	return !e.Used && !e.IsExpired()
}
