package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/models"
	"gorm.io/gorm"
)

// ReviewStatsRepository handles review statistics data operations
type ReviewStatsRepository struct {
	db *gorm.DB
}

// NewReviewStatsRepository creates a new ReviewStatsRepository
func NewReviewStatsRepository(db *gorm.DB) *ReviewStatsRepository {
	return &ReviewStatsRepository{db: db}
}

// Create creates new review stats record
func (r *ReviewStatsRepository) Create(ctx context.Context, reviewID uuid.UUID) (*models.ReviewStats, error) {
	stats := &models.ReviewStats{
		ReviewID: reviewID,
		Views:    0,
		Likes:    0,
		Dislikes: 0,
	}

	if err := r.db.WithContext(ctx).Create(stats).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

// GetByReviewID gets review stats by review ID
func (r *ReviewStatsRepository) GetByReviewID(ctx context.Context, reviewID uuid.UUID) (*models.ReviewStats, error) {
	var stats models.ReviewStats
	err := r.db.WithContext(ctx).Where("review_id = ?", reviewID).First(&stats).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果不存在，创建一个新的
			return r.Create(ctx, reviewID)
		}
		return nil, err
	}
	return &stats, nil
}

// IncrementViews increments the view count for a review
func (r *ReviewStatsRepository) IncrementViews(ctx context.Context, reviewID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.ReviewStats{}).
		Where("review_id = ?", reviewID).
		UpdateColumn("views", gorm.Expr("views + ?", 1))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		stats := &models.ReviewStats{
			ReviewID: reviewID,
			Views:    1,
		}

		err := r.db.WithContext(ctx).Create(stats).Error
		if err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return r.IncrementViews(ctx, reviewID)
			}
			return err
		}
	}

	return nil
}

// IncrementLikes increments the like count for a review
func (r *ReviewStatsRepository) IncrementLikes(ctx context.Context, reviewID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.ReviewStats{}).
		Where("review_id = ?", reviewID).
		UpdateColumn("likes", gorm.Expr("likes + ?", 1)).Error
}

// DecrementLikes decrements the like count for a review
func (r *ReviewStatsRepository) DecrementLikes(ctx context.Context, reviewID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.ReviewStats{}).
		Where("review_id = ?", reviewID).
		UpdateColumn("likes", gorm.Expr("likes - ?", 1)).Error
}

// IncrementDislikes increments the dislike count for a review
func (r *ReviewStatsRepository) IncrementDislikes(ctx context.Context, reviewID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.ReviewStats{}).
		Where("review_id = ?", reviewID).
		UpdateColumn("dislikes", gorm.Expr("dislikes + ?", 1)).Error
}

// DecrementDislikes decrements the dislike count for a review
func (r *ReviewStatsRepository) DecrementDislikes(ctx context.Context, reviewID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.ReviewStats{}).
		Where("review_id = ?", reviewID).
		UpdateColumn("dislikes", gorm.Expr("dislikes - ?", 1)).Error
}

// ReviewReactionRepository handles user reactions to reviews
type ReviewReactionRepository struct {
	db *gorm.DB
}

// NewReviewReactionRepository creates a new ReviewReactionRepository
func NewReviewReactionRepository(db *gorm.DB) *ReviewReactionRepository {
	return &ReviewReactionRepository{db: db}
}

// Create creates a new reaction record
func (r *ReviewReactionRepository) Create(ctx context.Context, reaction *models.ReviewReaction) error {
	return r.db.WithContext(ctx).Create(reaction).Error
}

// GetByReviewAndUser gets a reaction by review ID and user ID
func (r *ReviewReactionRepository) GetByReviewAndUser(ctx context.Context, reviewID, userID uuid.UUID) (*models.ReviewReaction, error) {
	var reaction models.ReviewReaction
	err := r.db.WithContext(ctx).
		Where("review_id = ? AND user_id = ?", reviewID, userID).
		First(&reaction).Error
	if err != nil {
		return nil, err
	}
	return &reaction, nil
}

// Delete deletes a reaction record
func (r *ReviewReactionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.ReviewReaction{}, "id = ?", id).Error
}

// Update updates a reaction record
func (r *ReviewReactionRepository) Update(ctx context.Context, reaction *models.ReviewReaction) error {
	return r.db.WithContext(ctx).Save(reaction).Error
}

// SiteStatsRepository handles site-wide statistics
type SiteStatsRepository struct {
	db *gorm.DB
}

// NewSiteStatsRepository creates a new SiteStatsRepository
func NewSiteStatsRepository(db *gorm.DB) *SiteStatsRepository {
	return &SiteStatsRepository{db: db}
}

// GetOrCreate gets or creates site stats record
func (r *SiteStatsRepository) GetOrCreate(ctx context.Context) (*models.SiteStats, error) {
	var stats models.SiteStats
	err := r.db.WithContext(ctx).First(&stats).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新的站点统计记录
			stats = models.SiteStats{
				TotalViews: 0,
			}
			if err := r.db.WithContext(ctx).Create(&stats).Error; err != nil {
				return nil, err
			}
			return &stats, nil
		}
		return nil, err
	}
	return &stats, nil
}

// IncrementTotalViews increments the total site views
func (r *SiteStatsRepository) IncrementTotalViews(ctx context.Context) error {
	result := r.db.WithContext(ctx).
		Model(&models.SiteStats{}).
		UpdateColumn("total_views", gorm.Expr("total_views + ?", 1))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		stats := &models.SiteStats{
			TotalViews: 1,
		}

		err := r.db.WithContext(ctx).Create(stats).Error
		if err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return r.IncrementTotalViews(ctx)
			}
			return err
		}
	}

	return nil
}

// GetTotalViews gets the total site views
func (r *SiteStatsRepository) GetTotalViews(ctx context.Context) (int64, error) {
	var stats models.SiteStats
	err := r.db.WithContext(ctx).First(&stats).Error
	if err != nil {
		return 0, err
	}
	return stats.TotalViews, nil
}
