package admin

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/repository"
	"gorm.io/gorm"
)

// UserAdminHandler exposes admin operations for user management.
type UserAdminHandler struct {
	users *repository.UserRepository
}

// NewUserAdminHandler constructs a UserAdminHandler.
func NewUserAdminHandler(users *repository.UserRepository) *UserAdminHandler {
	return &UserAdminHandler{users: users}
}

// List returns paginated users for admin view.
func (h *UserAdminHandler) List(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil || pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	users, err := h.users.List(offset, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
		return
	}

	total, err := h.users.Count()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户总数失败"})
		return
	}

	resp := make([]gin.H, 0, len(users))
	for _, user := range users {
		resp = append(resp, gin.H{
			"id":                user.ID,
			"email":             user.Email,
			"display_name":      user.DisplayName,
			"role":              user.Role,
			"email_verified":    user.EmailVerified,
			"email_verified_at": user.EmailVerifiedAt,
			"created_at":        user.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": resp,
		"pagination": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

// Delete removes a user by id.
func (h *UserAdminHandler) Delete(c *gin.Context) {
	userIDParam := c.Param("id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	currentUserID := c.MustGet("user_id").(uuid.UUID)
	if currentUserID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能删除当前登录的账户"})
		return
	}

	if _, err := h.users.FindByID(userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询用户失败"})
		return
	}

	if err := h.users.Delete(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户已删除"})
}
