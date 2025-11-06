package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents an application account.
type User struct {
	ID           uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;size:255;not null" json:"email"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	DisplayName  string    `gorm:"size:100;not null" json:"display_name"`
	Role         string    `gorm:"size:20;default:user" json:"role"`
	EmailVerified bool      `gorm:"default:false" json:"email_verified"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Reviews      []Review  `gorm:"foreignKey:AuthorID" json:"-"`
}

// BeforeCreate hook to set UUIDs automatically.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
