package admin

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hdu-dp/backend/internal/httpx"
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
	page := httpx.QueryInt(c, "page", 1, 1, 0)
	pageSize := httpx.QueryInt(c, "page_size", 20, 1, 100)

	offset := (page - 1) * pageSize

	users, err := h.users.List(offset, pageSize)
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, "获取用户列表失败")
		return
	}

	total, err := h.users.Count()
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, "获取用户总数失败")
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
	userID, ok := httpx.ParamUUID(c, "id", "无效的用户ID")
	if !ok {
		return
	}

	currentUserID, ok := httpx.MustContextUUID(c, "user_id", "missing user", "invalid user id")
	if !ok {
		return
	}
	if currentUserID == userID {
		httpx.Error(c, http.StatusBadRequest, "不能删除当前登录的账户")
		return
	}

	if _, err := h.users.FindByID(userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpx.Error(c, http.StatusNotFound, "用户不存在")
			return
		}
		httpx.Error(c, http.StatusInternalServerError, "查询用户失败")
		return
	}

	if err := h.users.Delete(userID); err != nil {
		httpx.Error(c, http.StatusInternalServerError, "删除用户失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户已删除"})
}
