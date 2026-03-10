package database

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hdu-dp/backend/internal/config"
	"github.com/hdu-dp/backend/internal/models"
	"github.com/hdu-dp/backend/internal/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newSeedTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbName := strings.ReplaceAll(t.Name(), "/", "_")
	dsn := fmt.Sprintf("file:%s_%d?mode=memory&cache=shared", dbName, time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	return db
}

func TestSeedAdminCreatesAdminWhenMissing(t *testing.T) {
	db := newSeedTestDB(t)
	cfg := &config.Config{}
	cfg.Admin.Email = "admin@example.com"
	cfg.Admin.Password = "Admin123!"

	if err := seedAdmin(db, cfg); err != nil {
		t.Fatalf("seedAdmin returned error: %v", err)
	}

	var admin models.User
	if err := db.Where("email = ?", cfg.Admin.Email).First(&admin).Error; err != nil {
		t.Fatalf("query created admin: %v", err)
	}

	if admin.Role != "admin" {
		t.Fatalf("expected role=admin, got %s", admin.Role)
	}
	if admin.DisplayName != "Administrator" {
		t.Fatalf("expected display name Administrator, got %s", admin.DisplayName)
	}
	if err := utils.CheckPassword(admin.PasswordHash, cfg.Admin.Password); err != nil {
		t.Fatalf("expected stored password to match seed password: %v", err)
	}
}

func TestSeedAdminUpdatesExistingUser(t *testing.T) {
	db := newSeedTestDB(t)
	cfg := &config.Config{}
	cfg.Admin.Email = "admin@example.com"
	cfg.Admin.Password = "NewAdmin123!"

	oldHash, err := utils.HashPassword("old-password")
	if err != nil {
		t.Fatalf("hash old password: %v", err)
	}

	existing := &models.User{
		Email:        cfg.Admin.Email,
		PasswordHash: oldHash,
		DisplayName:  "",
		Role:         "user",
	}
	if err := db.Create(existing).Error; err != nil {
		t.Fatalf("create existing user: %v", err)
	}

	if err := seedAdmin(db, cfg); err != nil {
		t.Fatalf("seedAdmin returned error: %v", err)
	}

	var updated models.User
	if err := db.Where("email = ?", cfg.Admin.Email).First(&updated).Error; err != nil {
		t.Fatalf("query updated admin: %v", err)
	}

	if updated.Role != "admin" {
		t.Fatalf("expected role=admin, got %s", updated.Role)
	}
	if updated.DisplayName != "Administrator" {
		t.Fatalf("expected display name Administrator, got %s", updated.DisplayName)
	}
	if err := utils.CheckPassword(updated.PasswordHash, cfg.Admin.Password); err != nil {
		t.Fatalf("expected stored password to be rotated: %v", err)
	}
}
