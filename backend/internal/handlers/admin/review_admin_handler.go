package admin

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/services"
)

// ReviewAdminHandler contains endpoints reserved for administrators.
type ReviewAdminHandler struct {
	reviews *services.ReviewService
}

// NewReviewAdminHandler constructs a new handler.
func NewReviewAdminHandler(reviews *services.ReviewService) *ReviewAdminHandler {
	return &ReviewAdminHandler{reviews: reviews}
}

// @Summary      待审核点评列表
// @Description  获取等待管理员审核的点评列表，支持分页、搜索和排序。
// @Tags         管理
// @Produce      json
// @Param        page      query int    false "页码" default(1)
// @Param        page_size query int    false "每页数量" default(10)
// @Param        query     query string false "搜索关键词"
// @Param        sort      query string false "排序字段 (created_at, rating)" enums(created_at, rating) default(created_at)
// @Param        order     query string false "排序顺序 (asc, desc)" enums(asc, desc) default(desc)
// @Success      200 {object} services.ReviewListResult
// @Failure      500 {object} object{error=string} "服务器内部错误"
// @Security     ApiKeyAuth
// @Router       /admin/reviews/pending [get]
func (h *ReviewAdminHandler) Pending(c *gin.Context) {
	filters := services.ListFilters{
		Page:     mustAtoi(c.DefaultQuery("page", "1")),
		PageSize: mustAtoi(c.DefaultQuery("page_size", "10")),
		Query:    strings.TrimSpace(c.Query("query")),
		SortBy:   c.DefaultQuery("sort", "created_at"),
		SortDir:  c.DefaultQuery("order", "desc"),
	}

	result, err := h.reviews.ListPending(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// @Summary      批准点评
// @Description  将指定 ID 的点评状态标记为“已批准”。
// @Tags         管理
// @Produce      json
// @Param        id path string true "点评 ID"
// @Success      200 {object} models.Review "批准成功"
// @Failure      400 {object} object{error=string} "无效的点评 ID 或状态错误"
// @Failure      404 {object} object{error=string} "点评不存在"
// @Security     ApiKeyAuth
// @Router       /admin/reviews/{id}/approve [put]
func (h *ReviewAdminHandler) Approve(c *gin.Context) {
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

	if err := h.reviews.Approve(review); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, review)
}

// @Summary      拒绝点评
// @Description  将指定 ID 的点评状态标记为“已拒绝”，并记录原因。
// @Tags         管理
// @Accept       json
// @Produce      json
// @Param        id   path string true "点评 ID"
// @Param        body body object{reason=string} true "拒绝原因"
// @Success      200  {object} models.Review "拒绝成功"
// @Failure      400  {object} object{error=string} "无效的点评 ID 或请求参数错误"
// @Failure      404  {object} object{error=string} "点评不存在"
// @Security     ApiKeyAuth
// @Router       /admin/reviews/{id}/reject [put]
func (h *ReviewAdminHandler) Reject(c *gin.Context) {
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

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if err := h.reviews.Reject(review, req.Reason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, review)
}

// @Summary      删除点评
// @Description  永久删除指定 ID 的点评及其关联的图片。
// @Tags         管理
// @Produce      json
// @Param        id path string true "点评 ID"
// @Success      204 "删除成功"
// @Failure      400 {object} object{error=string} "无效的点评 ID"
// @Failure      404 {object} object{error=string} "点评不存在"
// @Failure      500 {object} object{error=string} "服务器内部错误"
// @Security     ApiKeyAuth
// @Router       /admin/reviews/{id} [delete]
func (h *ReviewAdminHandler) Delete(c *gin.Context) {
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

	if err := h.reviews.DeleteReview(c.Request.Context(), review); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func mustAtoi(val string) int {
	n, _ := strconv.Atoi(val)
	return n
}
