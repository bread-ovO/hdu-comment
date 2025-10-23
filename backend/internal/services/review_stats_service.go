package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/models"
	"github.com/hdu-dp/backend/internal/repository"
	"gorm.io/gorm"
)

// ReviewStatsService handles review statistics business logic
type ReviewStatsService struct {
	reviewStatsRepo    *repository.ReviewStatsRepository
	reviewReactionRepo *repository.ReviewReactionRepository
	siteStatsRepo      *repository.SiteStatsRepository
}

// NewReviewStatsService creates a new ReviewStatsService
func NewReviewStatsService(
	reviewStatsRepo *repository.ReviewStatsRepository,
	reviewReactionRepo *repository.ReviewReactionRepository,
	siteStatsRepo *repository.SiteStatsRepository,
) *ReviewStatsService {
	return &ReviewStatsService{
		reviewStatsRepo:    reviewStatsRepo,
		reviewReactionRepo: reviewReactionRepo,
		siteStatsRepo:      siteStatsRepo,
	}
}

// GetReviewStats gets statistics for a specific review
func (s *ReviewStatsService) GetReviewStats(ctx context.Context, reviewID uuid.UUID) (*models.ReviewStats, error) {
	return s.reviewStatsRepo.GetByReviewID(ctx, reviewID)
}

// RecordView records a view for a review and increments site total views
func (s *ReviewStatsService) RecordView(ctx context.Context, reviewID uuid.UUID) error {
	// 增加点评浏览量
	if err := s.reviewStatsRepo.IncrementViews(ctx, reviewID); err != nil {
		return err
	}

	// 增加网站总浏览量
	return s.siteStatsRepo.IncrementTotalViews(ctx)
}

// GetUserReaction gets the user's reaction to a review
func (s *ReviewStatsService) GetUserReaction(ctx context.Context, reviewID, userID uuid.UUID) (*models.ReviewReaction, error) {
	return s.reviewReactionRepo.GetByReviewAndUser(ctx, reviewID, userID)
}

// ToggleReaction toggles a user's reaction to a review
func (s *ReviewStatsService) ToggleReaction(ctx context.Context, reviewID, userID uuid.UUID, reactionType models.ReactionType) error {
	// 获取现有反应
	existingReaction, err := s.reviewReactionRepo.GetByReviewAndUser(ctx, reviewID, userID)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// 如果已有反应
	if existingReaction != nil {
		// 如果反应类型相同，则取消反应
		if existingReaction.Type == reactionType {
			// 删除反应
			if err := s.reviewReactionRepo.Delete(ctx, existingReaction.ID); err != nil {
				return err
			}

			// 减少相应的计数
			if reactionType == models.ReactionTypeLike {
				return s.reviewStatsRepo.DecrementLikes(ctx, reviewID)
			} else {
				return s.reviewStatsRepo.DecrementDislikes(ctx, reviewID)
			}
		} else {
			// 反应类型不同，更新反应
			existingReaction.Type = reactionType
			if err := s.reviewReactionRepo.Update(ctx, existingReaction); err != nil {
				return err
			}

			// 更新计数：减少旧类型，增加新类型
			if reactionType == models.ReactionTypeLike {
				if err := s.reviewStatsRepo.DecrementDislikes(ctx, reviewID); err != nil {
					return err
				}
				return s.reviewStatsRepo.IncrementLikes(ctx, reviewID)
			} else {
				if err := s.reviewStatsRepo.DecrementLikes(ctx, reviewID); err != nil {
					return err
				}
				return s.reviewStatsRepo.IncrementDislikes(ctx, reviewID)
			}
		}
	} else {
		// 创建新反应
		newReaction := &models.ReviewReaction{
			ReviewID: reviewID,
			UserID:   userID,
			Type:     reactionType,
		}

		if err := s.reviewReactionRepo.Create(ctx, newReaction); err != nil {
			return err
		}

		// 增加相应的计数
		if reactionType == models.ReactionTypeLike {
			return s.reviewStatsRepo.IncrementLikes(ctx, reviewID)
		} else {
			return s.reviewStatsRepo.IncrementDislikes(ctx, reviewID)
		}
	}
}

// GetSiteStats gets site-wide statistics
func (s *ReviewStatsService) GetSiteStats(ctx context.Context) (*models.SiteStats, error) {
	return s.siteStatsRepo.GetOrCreate(ctx)
}

// GetSiteTotalViews gets the total site views
func (s *ReviewStatsService) GetSiteTotalViews(ctx context.Context) (int64, error) {
	return s.siteStatsRepo.GetTotalViews(ctx)
}

// GetReviewWithStats gets review with its statistics
type ReviewWithStats struct {
	*models.Review
	Stats        *models.ReviewStats    `json:"stats"`
	UserReaction *models.ReviewReaction `json:"user_reaction,omitempty"`
}

// GetReviewWithStats gets a review with its statistics and user reaction
func (s *ReviewStatsService) GetReviewWithStats(ctx context.Context, reviewID, userID uuid.UUID) (*ReviewWithStats, error) {
	// 这里需要从review service获取review，但为了解耦，我们假设调用者已经获取了review
	// 实际使用时，应该在review service中集成这些功能

	stats, err := s.GetReviewStats(ctx, reviewID)
	if err != nil {
		return nil, err
	}

	var userReaction *models.ReviewReaction
	if userID != uuid.Nil {
		userReaction, _ = s.GetUserReaction(ctx, reviewID, userID)
	}

	return &ReviewWithStats{
		Stats:        stats,
		UserReaction: userReaction,
	}, nil
}
