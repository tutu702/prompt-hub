// Package model - TuneHistory 调优历史表（ph_tune_histories）。
//
// 一次"调优诉求 → AI 改写"结果存一条；可由用户后续应用回原 prompt。
// 设计要点：
//   - (user_id, prompt_id, version_no) 唯一（同一用户同一提示词下版本号不重）。
//   - prompt_id 删除由 service 层校验（不依赖 DB 外键，避免迁移复杂度）。
//   - model_config_id 本期保留字段但允许 NULL（P08 再接 LLM 时填充）。
//   - request_text / original_content / tuned_content 用 TEXT，service 层做长度校验。
package model

import (
	"time"

	"gorm.io/gorm"
)

// TuneHistory 调优历史表模型。
type TuneHistory struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	UserID          uint           `gorm:"not null;uniqueIndex:idx_user_prompt_version,priority:1" json:"user_id"`
	PromptID        uint           `gorm:"not null;uniqueIndex:idx_user_prompt_version,priority:2" json:"prompt_id"`
	VersionNo       int            `gorm:"not null;uniqueIndex:idx_user_prompt_version,priority:3" json:"version_no"`
	RequestText     string         `gorm:"type:text;not null" json:"request_text"`
	OriginalContent string         `gorm:"type:text;not null" json:"original_content"`
	TunedContent    string         `gorm:"type:text;not null" json:"tuned_content"`
	ModelConfigID   *uint          `gorm:"index" json:"model_config_id,omitempty"`
	IsApplied       bool           `gorm:"default:false;not null;index" json:"is_applied"`
}

// TableName 指定表名遵循 PRD 约定 ph_ 前缀。
func (TuneHistory) TableName() string {
	return "ph_tune_histories"
}