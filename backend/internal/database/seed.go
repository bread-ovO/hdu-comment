package database

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/config"
	"github.com/hdu-dp/backend/internal/models"
	"github.com/hdu-dp/backend/internal/utils"
	"gorm.io/gorm"
)

func seedAdmin(db *gorm.DB, cfg *config.Config) error {
	if cfg.Admin.Email == "" || cfg.Admin.Password == "" {
		return nil
	}

	var existing models.User
	err := db.Where("email = ?", cfg.Admin.Email).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("query admin: %w", err)
	}

	hashed, err := utils.HashPassword(cfg.Admin.Password)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}

	if err == nil {
		needsUpdate := false
		if existing.Role != "admin" {
			existing.Role = "admin"
			needsUpdate = true
		}
		if err := utils.CheckPassword(existing.PasswordHash, cfg.Admin.Password); err != nil {
			existing.PasswordHash = hashed
			needsUpdate = true
		}
		if existing.DisplayName == "" {
			existing.DisplayName = "Administrator"
			needsUpdate = true
		}
		if needsUpdate {
			if err := db.Save(&existing).Error; err != nil {
				return fmt.Errorf("update admin: %w", err)
			}
		}
		return nil
	}

	admin := models.User{
		ID:           uuid.New(),
		Email:        cfg.Admin.Email,
		PasswordHash: hashed,
		DisplayName:  "Administrator",
		Role:         "admin",
	}

	if err := db.Create(&admin).Error; err != nil {
		return fmt.Errorf("create admin: %w", err)
	}

	return nil
}
