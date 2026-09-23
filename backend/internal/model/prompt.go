// Package model - Prompt 提示词表（ph_prompts）。
//
// 系统核心实体；承载用户管理的提示词资产。
// 设计要点：
//   - user_id 强制隔离；任何读写必须带 user_id 过滤（repository 层负责）。
//   - category_id 可空（不分类时为 NULL）；删除分类时不级联，由 service 层阻断。
//   - tags 字段本期用 string 存 JSON 字符串；service 层负责与 []string 互转。
//   - version_no 用于 AI 调优时递增，标识当前内容属于第几版。
package model

import (
	"time"

	"gorm.io/gorm"
)

// Prompt 提示词模型。
type Prompt struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	UserID      uint           `gorm:"not null;index" json:"user_id"`
	CategoryID  *uint          `gorm:"index" json:"category_id,omitempty"`
	Title       string         `gorm:"size:100;not null" json:"title"`
	Description string         `gorm:"size:200;default:''" json:"description"`
	Content     string         `gorm:"type:text;not null" json:"content"`
	IsFavorite  bool           `gorm:"default:false;not null;index" json:"is_favorite"`
	// Tags 存 JSON 字符串（如 `["spinner","中文"]`），service 层负责与 []string 互转；
	// 选 string 而非 datatypes.JSON 以避免引入额外的 GORM 子包。
	Tags      string `gorm:"type:text;default:'[]'" json:"-"`
	VersionNo int    `gorm:"default:1;not null" json:"version_no"`
}

// TableName 指定表名遵循 PRD 约定 ph_ 前缀。
func (p *Prompt) TableName() string {
	return "ph_prompts"
}