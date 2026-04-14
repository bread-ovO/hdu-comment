package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/httpx"
	"github.com/hdu-dp/backend/internal/models"
	"github.com/hdu-dp/backend/internal/services"
	"github.com/hdu-dp/backend/internal/storage"
)

// ReviewHandler manages review related HTTP endpoints.
type ReviewHandler struct {
	reviews *services.ReviewService
}

const maxImageUploadSize = 10 * 1024 * 1024 // 10MB

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
		httpx.Error(c, http.StatusInternalServerError, err.Error())
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
	userID, ok := httpx.MustContextUUID(c, "user_id", "missing user", "invalid user id")
	if !ok {
		return
	}
	var req struct {
		Title       string  `json:"title" binding:"required,max=128"`
		Address     string  `json:"address" binding:"required,max=255"`
		Description string  `json:"description" binding:"max=4000"`
		Rating      float32 `json:"rating" binding:"required"`
	}

	if !httpx.BindJSON(c, &req, "请输入完整且有效的点评信息") {
		return
	}

	review, err := h.reviews.Submit(userID, services.CreateReviewInput{
		Title:       req.Title,
		Address:     req.Address,
		Description: req.Description,
		Rating:      req.Rating,
	})
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, err.Error())
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
	id, ok := httpx.ParamUUID(c, "id", "invalid review id")
	if !ok {
		return
	}

	review, err := h.reviews.Get(id)
	if err != nil {
		httpx.Error(c, http.StatusNotFound, "review not found")
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
				httpx.Error(c, http.StatusForbidden, "review not accessible")
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
	userID, ok := httpx.MustContextUUID(c, "user_id", "missing user", "invalid user id")
	if !ok {
		return
	}
	filters := parseListFilters(c)
	result, err := h.reviews.ListByAuthor(userID, filters)
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, err.Error())
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
// @Failure      413  {object}  object{error=string} "图片过大"
// @Failure      500  {object}  object{error=string} "服务器内部错误"
// @Security     ApiKeyAuth
// @Router       /reviews/{id}/images [post]
func (h *ReviewHandler) UploadImage(c *gin.Context) {
	reviewID, ok := httpx.ParamUUID(c, "id", "invalid review id")
	if !ok {
		return
	}

	review, err := h.reviews.Get(reviewID)
	if err != nil {
		httpx.Error(c, http.StatusNotFound, "review not found")
		return
	}

	userID, ok := httpx.MustContextUUID(c, "user_id", "missing user", "invalid user id")
	if !ok {
		return
	}
	if err := services.ValidateOwnership(review, userID); err != nil {
		httpx.Error(c, http.StatusForbidden, "not owner")
		return
	}
	if review.Status != models.ReviewStatusPending {
		httpx.Error(c, http.StatusBadRequest, "images can only be uploaded while review is pending")
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "file is required")
		return
	}
	if fileHeader.Size > maxImageUploadSize {
		httpx.Error(c, http.StatusRequestEntityTooLarge, "image size must not exceed 10MB")
		return
	}

	opened, err := fileHeader.Open()
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, err.Error())
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
		httpx.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, image)
}

func parseListFilters(c *gin.Context) services.ListFilters {
	query := strings.TrimSpace(c.Query("query"))
	sortBy := c.DefaultQuery("sort", "created_at")
	sortDir := c.DefaultQuery("order", "desc")

	return services.ListFilters{
		Page:     httpx.QueryInt(c, "page", 1, 1, 0),
		PageSize: httpx.QueryInt(c, "page_size", 10, 1, 100),
		Query:    query,
		SortBy:   sortBy,
		SortDir:  sortDir,
	}
}
