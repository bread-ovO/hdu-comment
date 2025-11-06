package repository

import (
	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/models"
	"gorm.io/gorm"
)

// UserRepository handles persistence for user entities.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository constructs a repository instance.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user entry.
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// FindByEmail fetches a user by email.
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID fetches a user by primary key.
func (r *UserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Save persists changes to a user.
func (r *UserRepository) Save(user *models.User) error {
	return r.db.Save(user).Error
}

// List returns users ordered by creation time desc with pagination.
func (r *UserRepository) List(offset, limit int) ([]models.User, error) {
	var users []models.User
	query := r.db.Order("created_at desc")
	if limit > 0 {
		query = query.Offset(offset).Limit(limit)
	}
	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// Count returns total count of users.
func (r *UserRepository) Count() (int64, error) {
	var total int64
	if err := r.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// Delete removes a user by id.
func (r *UserRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}
