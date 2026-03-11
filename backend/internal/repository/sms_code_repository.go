package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/models"
	"gorm.io/gorm"
)

// SMSCodeRepository handles sms code persistence.
type SMSCodeRepository struct {
	db *gorm.DB
}

// NewSMSCodeRepository creates a new SMSCodeRepository.
func NewSMSCodeRepository(db *gorm.DB) *SMSCodeRepository {
	return &SMSCodeRepository{db: db}
}

// Create inserts a new sms code.
func (r *SMSCodeRepository) Create(code *models.SMSCode) error {
	return r.db.Create(code).Error
}

// FindLatestActive finds the latest non-expired code for phone+purpose.
func (r *SMSCodeRepository) FindLatestActive(phone, purpose string, now time.Time) (*models.SMSCode, error) {
	var code models.SMSCode
	if err := r.db.
		Where("phone = ? AND purpose = ? AND used = ? AND expires_at > ?", phone, purpose, false, now).
		Order("created_at desc").
		First(&code).Error; err != nil {
		return nil, err
	}
	return &code, nil
}

// MarkUsed marks a code as used.
func (r *SMSCodeRepository) MarkUsed(id uuid.UUID) error {
	return r.db.Model(&models.SMSCode{}).Where("id = ?", id).Update("used", true).Error
}

// DeleteByPhonePurpose removes prior codes for a phone+purpose pair.
func (r *SMSCodeRepository) DeleteByPhonePurpose(phone, purpose string) error {
	return r.db.Where("phone = ? AND purpose = ?", phone, purpose).Delete(&models.SMSCode{}).Error
}

// DeleteExpired removes all expired or used records.
func (r *SMSCodeRepository) DeleteExpired(now time.Time) error {
	return r.db.Where("expires_at < ? OR used = ?", now, true).Delete(&models.SMSCode{}).Error
}
