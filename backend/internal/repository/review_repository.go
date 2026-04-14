package repository

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/models"
	"gorm.io/gorm"
)

// ReviewRepository manages persistence for reviews and images.
type ReviewRepository struct {
	db *gorm.DB
}

// NewReviewRepository constructs a review repository.
func NewReviewRepository(db *gorm.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

// ListOptions holds query parameters for retrieving reviews.
type ListOptions struct {
	Statuses []models.ReviewStatus
	AuthorID *uuid.UUID
	Query    string
	SortBy   string
	SortDir  string
	Limit    int
	Offset   int
}

// ListResult represents a paginated resultset.
type ListResult struct {
	Reviews []models.Review
	Total   int64
}

// List fetches reviews using provided options.
func (r *ReviewRepository) List(opts ListOptions) (ListResult, error) {
	base := r.db.Model(&models.Review{})

	if len(opts.Statuses) > 0 {
		base = base.Where("status IN ?", opts.Statuses)
	}
	if opts.AuthorID != nil {
		base = base.Where("author_id = ?", opts.AuthorID)
	}
	if opts.Query != "" {
		like := fmt.Sprintf("%%%s%%", opts.Query)
		base = base.Where("title LIKE ? OR address LIKE ? OR description LIKE ?", like, like, like)
	}

	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return ListResult{}, err
	}

	listQuery := base.Session(&gorm.Session{}).Preload("Images").Preload("Author")

	sortBy := "created_at"
	switch strings.ToLower(opts.SortBy) {
	case "rating":
		sortBy = "rating"
	case "created_at":
		sortBy = "created_at"
	}

	sortDir := "DESC"
	if strings.EqualFold(opts.SortDir, "asc") {
		sortDir = "ASC"
	}

	listQuery = listQuery.Order(fmt.Sprintf("%s %s", sortBy, sortDir))

	if opts.Limit > 0 {
		listQuery = listQuery.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		listQuery = listQuery.Offset(opts.Offset)
	}

	var reviews []models.Review
	if err := listQuery.Find(&reviews).Error; err != nil {
		return ListResult{}, err
	}

	return ListResult{Reviews: reviews, Total: total}, nil
}

// Create inserts a new review.
func (r *ReviewRepository) Create(review *models.Review) error {
	return r.db.Create(review).Error
}

// Update persists changes to a review.
func (r *ReviewRepository) Update(review *models.Review) error {
	return r.db.Save(review).Error
}

// FindByID returns a review by UUID including relations.
func (r *ReviewRepository) FindByID(id uuid.UUID) (*models.Review, error) {
	var review models.Review
	if err := r.db.Preload("Images").Preload("Author").First(&review, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &review, nil
}

// AddImage appends a review image entry.
func (r *ReviewRepository) AddImage(image *models.ReviewImage) error {
	return r.db.Create(image).Error
}

// DeleteImage removes an image by key.
func (r *ReviewRepository) DeleteImage(id uuid.UUID) error {
	return r.db.Delete(&models.ReviewImage{}, "id = ?", id).Error
}

// Delete removes a review and associated images inside a transaction.
func (r *ReviewRepository) Delete(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("review_id = ?", id).Delete(&models.ReviewReaction{}).Error; err != nil {
			return err
		}
		if err := tx.Where("review_id = ?", id).Delete(&models.ReviewStats{}).Error; err != nil {
			return err
		}
		if err := tx.Where("review_id = ?", id).Delete(&models.ReviewImage{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.Review{}, "id = ?", id).Error; err != nil {
			return err
		}
		return nil
	})
}
