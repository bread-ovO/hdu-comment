package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/models"
	"github.com/hdu-dp/backend/internal/services"
	"github.com/hdu-dp/backend/internal/storage"
)

// ReviewHandler manages review related HTTP endpoints.
type ReviewHandler struct {
	reviews *services.ReviewService
}

// NewReviewHandler constructs a ReviewHandler.
func NewReviewHandler(reviews *services.ReviewService) *ReviewHandler {
	return &ReviewHandler{reviews: reviews}
}

// @Summary      公开点评列表
// @Description  获取已审核通过的点评列表，支持分页、搜索和排序。
// @Tags         点评
// @Produce      json
// @Param        page      query int    false "页码" default(1)
// @Param        page_size query int    false "每页数量" default(10)
// @Param        query     query string false "搜索关键词"
// @Param        sort      query string false "排序字段 (created_at, rating)" enums(created_at, rating) default(created_at)
// @Param        order     query string false "排序顺序 (asc, desc)" enums(asc, desc) default(desc)
// @Success      200 {object} services.ReviewListResult
// @Failure      500 {object} object{error=string} "服务器内部错误"
// @Router       /reviews [get]
func (h *ReviewHandler) ListPublic(c *gin.Context) {
	filters := parseListFilters(c)
	result, err := h.reviews.ListPublic(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// @Summary      提交新点评
// @Description  已认证用户提交一条新的点评，需要等待管理员审核。
// @Tags         点评
// @Accept       json
// @Produce      json
// @Param        body body object{title=string,address=string,description=string,rating=number} true "点评内容"
// @Success      201 {object} models.Review "创建成功"
// @Failure      400 {object} object{error=string} "请求参数错误"
// @Security     ApiKeyAuth
// @Router       /reviews [post]
func (h *ReviewHandler) Submit(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	var req struct {
		Title       string  `json:"title"`
		Address     string  `json:"address"`
		Description string  `json:"description"`
		Rating      float32 `json:"rating"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	review, err := h.reviews.Submit(userID, services.CreateReviewInput{
		Title:       req.Title,
		Address:     req.Address,
		Description: req.Description,
		Rating:      req.Rating,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, review)
}

// @Summary      获取点评详情
// @Description  根据 ID 获取单个点评的详细信息。未审核的点评仅作者和管理员可见。
// @Tags         点评
// @Produce      json
// @Param        id path string true "点评 ID"
// @Success      200 {object} models.Review
// @Failure      400 {object} object{error=string} "无效的点评 ID"
// @Failure      403 {object} object{error=string} "无权访问"
// @Failure      404 {object} object{error=string} "点评不存在"
// @Router       /reviews/{id} [get]
func (h *ReviewHandler) Detail(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}

	review, err := h.reviews.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
		return
	}

	if review.Status != models.ReviewStatusApproved {
		roleVal, ok := c.Get("role")
		role := ""
		if ok {
			role, _ = roleVal.(string)
		}
		if role != "admin" {
			userVal, ok := c.Get("user_id")
			userID, okID := userVal.(uuid.UUID)
			if !ok || !okID || review.AuthorID != userID {
				c.JSON(http.StatusForbidden, gin.H{"error": "review not accessible"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, review)
}

// @Summary      我的点评列表
// @Description  获取当前认证用户提交的所有点评列表，支持分页、搜索和排序。
// @Tags         点评
// @Produce      json
// @Param        page      query int    false "页码" default(1)
// @Param        page_size query int    false "每页数量" default(10)
// @Param        query     query string false "搜索关键词"
// @Param        sort      query string false "排序字段 (created_at, rating)" enums(created_at, rating) default(created_at)
// @Param        order     query string false "排序顺序 (asc, desc)" enums(asc, desc) default(desc)
// @Success      200 {object} services.ReviewListResult
// @Failure      500 {object} object{error=string} "服务器内部错误"
// @Security     ApiKeyAuth
// @Router       /reviews/me [get]
func (h *ReviewHandler) MyReviews(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	filters := parseListFilters(c)
	result, err := h.reviews.ListByAuthor(userID, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// @Summary      上传点评图片
// @Description  为指定的点评上传一张图片。用户只能为自己的点评上传。
// @Tags         点评
// @Accept       multipart/form-data
// @Produce      json
// @Param        id   path      string true "点评 ID"
// @Param        file formData  file   true "图片文件"
// @Success      201  {object}  models.ReviewImage "上传成功"
// @Failure      400  {object}  object{error=string} "请求错误"
// @Failure      403  {object}  object{error=string} "无权操作"
// @Failure      404  {object}  object{error=string} "点评不存在"
// @Failure      500  {object}  object{error=string} "服务器内部错误"
// @Security     ApiKeyAuth
// @Router       /reviews/{id}/images [post]
func (h *ReviewHandler) UploadImage(c *gin.Context) {
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}

	review, err := h.reviews.Get(reviewID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	if err := services.ValidateOwnership(review, userID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "not owner"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	opened, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	uploadFile := &storage.UploadFile{
		Reader:      opened,
		Size:        fileHeader.Size,
		Filename:    fileHeader.Filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
	}

	image, err := h.reviews.StoreImage(c.Request.Context(), reviewID, uploadFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, image)
}

func parseListFilters(c *gin.Context) services.ListFilters {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	query := strings.TrimSpace(c.Query("query"))
	sortBy := c.DefaultQuery("sort", "created_at")
	sortDir := c.DefaultQuery("order", "desc")

	return services.ListFilters{
		Page:     page,
		PageSize: pageSize,
		Query:    query,
		SortBy:   sortBy,
		SortDir:  sortDir,
	}
}
