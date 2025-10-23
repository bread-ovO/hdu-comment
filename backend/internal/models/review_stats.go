package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReviewStats 存储点评的统计数据
type ReviewStats struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	ReviewID  uuid.UUID `gorm:"type:char(36);not null;uniqueIndex" json:"review_id"`
	Views     int64     `gorm:"default:0" json:"views"`
	Likes     int64     `gorm:"default:0" json:"likes"`
	Dislikes  int64     `gorm:"default:0" json:"dislikes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate assigns a UUID if empty.
func (rs *ReviewStats) BeforeCreate(tx *gorm.DB) error {
	if rs.ID == uuid.Nil {
		rs.ID = uuid.New()
	}
	return nil
}

// ReviewReaction 存储用户对点评的点赞/踩记录
type ReviewReaction struct {
	ID        uuid.UUID    `gorm:"type:char(36);primaryKey" json:"id"`
	ReviewID  uuid.UUID    `gorm:"type:char(36);not null;index" json:"review_id"`
	UserID    uuid.UUID    `gorm:"type:char(36);not null;index" json:"user_id"`
	Type      ReactionType `gorm:"size:10;not null" json:"type"` // like or dislike
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`

	// 复合唯一索引，确保用户对同一点评只能有一个反应
	_ struct{} `gorm:"uniqueIndex:idx_review_user;constraint:OnDelete:CASCADE"`
}

// ReactionType 定义反应类型
type ReactionType string

const (
	ReactionTypeLike    ReactionType = "like"
	ReactionTypeDislike ReactionType = "dislike"
)

// BeforeCreate assigns a UUID if empty.
func (rr *ReviewReaction) BeforeCreate(tx *gorm.DB) error {
	if rr.ID == uuid.Nil {
		rr.ID = uuid.New()
	}
	return nil
}

// SiteStats 存储网站整体统计数据
type SiteStats struct {
	ID         uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	TotalViews int64     `gorm:"default:0" json:"total_views"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// BeforeCreate assigns a UUID if empty.
func (ss *SiteStats) BeforeCreate(tx *gorm.DB) error {
	if ss.ID == uuid.Nil {
		ss.ID = uuid.New()
	}
	return nil
}
