package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/models"
	"github.com/hdu-dp/backend/internal/services"
)

// ReviewStatsHandler handles review statistics HTTP endpoints
type ReviewStatsHandler struct {
	reviewStatsService *services.ReviewStatsService
}

// NewReviewStatsHandler creates a new ReviewStatsHandler
func NewReviewStatsHandler(reviewStatsService *services.ReviewStatsService) *ReviewStatsHandler {
	return &ReviewStatsHandler{
		reviewStatsService: reviewStatsService,
	}
}

// @Summary      获取点评统计信息
// @Description  获取指定点评的浏览量、点赞数、踩数等统计信息
// @Tags         点评统计
// @Produce      json
// @Param        id path string true "点评 ID"
// @Success      200 {object} models.ReviewStats
// @Failure      400 {object} object{error=string} "无效的点评 ID"
// @Failure      500 {object} object{error=string} "服务器内部错误"
// @Router       /reviews/{id}/stats [get]
func (h *ReviewStatsHandler) GetReviewStats(c *gin.Context) {
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}

	stats, err := h.reviewStatsService.GetReviewStats(c.Request.Context(), reviewID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// @Summary      记录点评浏览
// @Description  记录一次点评浏览，增加浏览量
// @Tags         点评统计
// @Produce      json
// @Param        id path string true "点评 ID"
// @Success      200 {object} object{message=string}
// @Failure      400 {object} object{error=string} "无效的点评 ID"
// @Failure      500 {object} object{error=string} "服务器内部错误"
// @Router       /reviews/{id}/view [post]
func (h *ReviewStatsHandler) RecordView(c *gin.Context) {
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}

	if err := h.reviewStatsService.RecordView(c.Request.Context(), reviewID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "view recorded"})
}

// @Summary      点赞或踩点评
// @Description  对点评进行点赞或踩操作，再次点击相同类型会取消操作
// @Tags         点评统计
// @Accept       json
// @Produce      json
// @Param        id path string true "点评 ID"
// @Param        body body object{type=string} true "反应类型 (like 或 dislike)"
// @Success      200 {object} object{message=string,action=string}
// @Failure      400 {object} object{error=string} "请求参数错误"
// @Failure      401 {object} object{error=string} "未认证"
// @Failure      500 {object} object{error=string} "服务器内部错误"
// @Security     ApiKeyAuth
// @Router       /reviews/{id}/react [post]
func (h *ReviewStatsHandler) ToggleReaction(c *gin.Context) {
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	var req struct {
		Type string `json:"type" binding:"required,oneof=like dislike"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reactionType := models.ReactionType(req.Type)
	if err := h.reviewStatsService.ToggleReaction(c.Request.Context(), reviewID, userID, reactionType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "reaction updated",
		"action":  req.Type,
	})
}

// @Summary      获取用户对点评的反应
// @Description  获取当前用户对指定点评的点赞/踩状态
// @Tags         点评统计
// @Produce      json
// @Param        id path string true "点评 ID"
// @Success      200 {object} object{reaction=string}
// @Failure      400 {object} object{error=string} "无效的点评 ID"
// @Failure      401 {object} object{error=string} "未认证"
// @Failure      500 {object} object{error=string} "服务器内部错误"
// @Security     ApiKeyAuth
// @Router       /reviews/{id}/user-reaction [get]
func (h *ReviewStatsHandler) GetUserReaction(c *gin.Context) {
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	reaction, err := h.reviewStatsService.GetUserReaction(c.Request.Context(), reviewID, userID)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusOK, gin.H{"reaction": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reaction": reaction.Type})
}

// @Summary      获取网站统计信息
// @Description  获取网站总浏览量等统计信息
// @Tags         网站统计
// @Produce      json
// @Success      200 {object} models.SiteStats
// @Failure      500 {object} object{error=string} "服务器内部错误"
// @Router       /stats/site [get]
func (h *ReviewStatsHandler) GetSiteStats(c *gin.Context) {
	stats, err := h.reviewStatsService.GetSiteStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// @Summary      获取网站总浏览量
// @Description  获取网站总浏览量
// @Tags         网站统计
// @Produce      json
// @Success      200 {object} object{total_views=int64}
// @Failure      500 {object} object{error=string} "服务器内部错误"
// @Router       /stats/total-views [get]
func (h *ReviewStatsHandler) GetTotalViews(c *gin.Context) {
	totalViews, err := h.reviewStatsService.GetSiteTotalViews(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"total_views": totalViews})
}
