package services

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/common"
	"github.com/hdu-dp/backend/internal/models"
	"github.com/hdu-dp/backend/internal/repository"
	"github.com/hdu-dp/backend/internal/storage"
)

// ReviewService contains business logic around review workflows.
type ReviewService struct {
	reviews *repository.ReviewRepository
	storage storage.FileStorage
}

// NewReviewService constructs a review service instance.
func NewReviewService(reviews *repository.ReviewRepository, fileStorage storage.FileStorage) *ReviewService {
	return &ReviewService{reviews: reviews, storage: fileStorage}
}

// CreateReviewInput bundles parameters for a new review.
type CreateReviewInput struct {
	Title       string
	Address     string
	Description string
	Rating      float32
}

// ListFilters describes filters sortable/paginatable lists.
type ListFilters struct {
	Page     int
	PageSize int
	Query    string
	SortBy   string
	SortDir  string
}

// Pagination metadata for list responses.
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ReviewListResult wraps review list responses with pagination info.
type ReviewListResult struct {
	Data       []models.Review `json:"data"`
	Pagination Pagination      `json:"pagination"`
}

// Submit creates a new review in pending state.
func (s *ReviewService) Submit(authorID uuid.UUID, input CreateReviewInput) (*models.Review, error) {
	title := strings.TrimSpace(input.Title)
	address := strings.TrimSpace(input.Address)
	description := strings.TrimSpace(input.Description)

	if title == "" || address == "" {
		return nil, errors.New("title and address are required")
	}
	if input.Rating < 0 || input.Rating > 5 {
		return nil, errors.New("rating must be between 0 and 5")
	}

	review := &models.Review{
		ID:          uuid.New(),
		Title:       title,
		Address:     address,
		Description: description,
		Rating:      input.Rating,
		Status:      models.ReviewStatusPending,
		AuthorID:    authorID,
	}

	if err := s.reviews.Create(review); err != nil {
		return nil, err
	}
	return review, nil
}

// ListPublic returns approved reviews.
func (s *ReviewService) ListPublic(filters ListFilters) (ReviewListResult, error) {
	opts := buildListOptions(filters)
	opts.Statuses = []models.ReviewStatus{models.ReviewStatusApproved}
	return s.listWithPagination(opts, filters)
}

// ListByAuthor returns reviews submitted by the specified user.
func (s *ReviewService) ListByAuthor(authorID uuid.UUID, filters ListFilters) (ReviewListResult, error) {
	opts := buildListOptions(filters)
	opts.AuthorID = &authorID
	return s.listWithPagination(opts, filters)
}

// ListPending returns pending reviews for admin review.
func (s *ReviewService) ListPending(filters ListFilters) (ReviewListResult, error) {
	opts := buildListOptions(filters)
	opts.Statuses = []models.ReviewStatus{models.ReviewStatusPending}
	return s.listWithPagination(opts, filters)
}

func buildListOptions(filters ListFilters) repository.ListOptions {
	limit := filters.PageSize
	if limit <= 0 {
		limit = 10
	}
	page := filters.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	return repository.ListOptions{
		Query:   filters.Query,
		SortBy:  filters.SortBy,
		SortDir: filters.SortDir,
		Limit:   limit,
		Offset:  offset,
	}
}

func (s *ReviewService) listWithPagination(opts repository.ListOptions, filters ListFilters) (ReviewListResult, error) {
	result, err := s.reviews.List(opts)
	if err != nil {
		return ReviewListResult{}, err
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 10
	}
	page := filters.Page
	if page <= 0 {
		page = 1
	}

	totalPages := int((result.Total + int64(limit) - 1) / int64(limit))

	return ReviewListResult{
		Data: result.Reviews,
		Pagination: Pagination{
			Page:       page,
			PageSize:   limit,
			Total:      result.Total,
			TotalPages: totalPages,
		},
	}, nil
}

// Get returns a review by ID.
func (s *ReviewService) Get(id uuid.UUID) (*models.Review, error) {
	return s.reviews.FindByID(id)
}

// Approve marks a review as approved.
func (s *ReviewService) Approve(review *models.Review) error {
	if review.Status != models.ReviewStatusPending {
		return common.ErrReviewAlreadyProcessed
	}
	review.Status = models.ReviewStatusApproved
	review.RejectionReason = ""
	return s.reviews.Update(review)
}

// Reject marks a review as rejected with reason.
func (s *ReviewService) Reject(review *models.Review, reason string) error {
	if review.Status != models.ReviewStatusPending {
		return common.ErrReviewAlreadyProcessed
	}
	review.Status = models.ReviewStatusRejected
	review.RejectionReason = strings.TrimSpace(reason)
	return s.reviews.Update(review)
}

// StoreImage saves the uploaded file via storage provider and records metadata.
func (s *ReviewService) StoreImage(ctx context.Context, reviewID uuid.UUID, file *storage.UploadFile) (*models.ReviewImage, error) {
	if file == nil {
		return nil, errors.New("file payload required")
	}

	defer file.Reader.Close()

	key := filepath.ToSlash(filepath.Join(reviewID.String(), fmt.Sprintf("%d_%s", time.Now().UnixNano(), sanitizeFilename(file.Filename))))

	info, err := s.storage.Save(ctx, key, file.Reader, file.Size, file.ContentType)
	if err != nil {
		return nil, err
	}

	image := &models.ReviewImage{
		ReviewID:   reviewID,
		StorageKey: info.Key,
		URL:        info.URL,
	}

	if err := s.reviews.AddImage(image); err != nil {
		return nil, err
	}

	return image, nil
}

// DeleteReview removes a review and attempts to clean up related assets.
func (s *ReviewService) DeleteReview(ctx context.Context, review *models.Review) error {
	if review == nil {
		return errors.New("review is required")
	}

	// 重新获取包含图片的完整评论数据
	fullReview, err := s.reviews.FindByID(review.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch review with images: %w", err)
	}

	var keys []string
	if len(fullReview.Images) > 0 {
		keys = make([]string, 0, len(fullReview.Images))
		for _, image := range fullReview.Images {
			if image.StorageKey != "" {
				keys = append(keys, image.StorageKey)
			}
		}
	}

	if err := s.reviews.Delete(review.ID); err != nil {
		return err
	}

	// 清理存储中的图片文件，但忽略错误以避免删除失败
	for _, key := range keys {
		_ = s.storage.Delete(ctx, key)
	}

	return nil
}

func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, " ", "_")
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '_' || r == '-' || r == '.':
			return r
		default:
			return '_'
		}
	}, name)
}

// ValidateOwnership ensures the review belongs to the user.
func ValidateOwnership(review *models.Review, userID uuid.UUID) error {
	if review.AuthorID != userID {
		return fmt.Errorf("review does not belong to user")
	}
	return nil
}
