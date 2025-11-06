package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/models"
	"gorm.io/gorm"
)

// EmailVerificationRepository handles email verification data access
type EmailVerificationRepository struct {
	db *gorm.DB
}

// NewEmailVerificationRepository creates a new EmailVerificationRepository
func NewEmailVerificationRepository(db *gorm.DB) *EmailVerificationRepository {
	return &EmailVerificationRepository{db: db}
}

// Create creates a new email verification record
func (r *EmailVerificationRepository) Create(ctx context.Context, verification *models.EmailVerification) error {
	return r.db.WithContext(ctx).Create(verification).Error
}

// GetByToken retrieves a verification record by token
func (r *EmailVerificationRepository) GetByToken(ctx context.Context, token string) (*models.EmailVerification, error) {
	var verification models.EmailVerification
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&verification).Error
	if err != nil {
		return nil, err
	}
	return &verification, nil
}

// GetByEmailAndToken retrieves a verification record by email and token
func (r *EmailVerificationRepository) GetByEmailAndToken(ctx context.Context, email, token string) (*models.EmailVerification, error) {
	var verification models.EmailVerification
	err := r.db.WithContext(ctx).Where("email = ? AND token = ?", email, token).First(&verification).Error
	if err != nil {
		return nil, err
	}
	return &verification, nil
}

// GetByUserID retrieves the latest verification record for a user
func (r *EmailVerificationRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.EmailVerification, error) {
	var verification models.EmailVerification
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").First(&verification).Error
	if err != nil {
		return nil, err
	}
	return &verification, nil
}

// MarkAsUsed marks a verification token as used
func (r *EmailVerificationRepository) MarkAsUsed(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Model(&models.EmailVerification{}).
		Where("token = ?", token).
		Update("used", true).Error
}

// LinkTokenToUser associates a verification token with a user and marks it as used
func (r *EmailVerificationRepository) LinkTokenToUser(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.EmailVerification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"user_id": userID,
			"used":    true,
		}).Error
}

// DeleteExpired deletes expired verification tokens
func (r *EmailVerificationRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ? OR used = ?", time.Now(), true).
		Delete(&models.EmailVerification{}).Error
}

// DeleteByUserID deletes all verification tokens for a user
func (r *EmailVerificationRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.EmailVerification{}).Error
}

// DeleteByEmail deletes all verification tokens for an email address
func (r *EmailVerificationRepository) DeleteByEmail(ctx context.Context, email string) error {
	return r.db.WithContext(ctx).
		Where("email = ?", email).
		Delete(&models.EmailVerification{}).Error
}
