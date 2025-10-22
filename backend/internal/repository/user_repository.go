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
