// Package model - LLMConfig 大模型配置表（ph_llm_configs）。
//
// 字段说明：
//   - Name：别名，1-20 字符，同一用户下唯一
//   - Provider：openai/anthropic/deepseek/qwen/zhipu/moonshot/custom
//   - APIKeyEnc：AES-GCM 加密后的密文（绝不直接给前端）
//   - APIKeyPrefix / APIKeySuffix：明文首 6 / 末 4，用于列表脱敏展示（避免列表解密）
//   - IsDefault：唯一（service 层事务保证）
package model

import (
	"time"

	"gorm.io/gorm"
)

// LLMConfig 大模型配置。
type LLMConfig struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID   uint   `gorm:"not null;index" json:"user_id"`
	Name     string `gorm:"size:20;not null" json:"name"`
	Provider string `gorm:"size:32;not null" json:"provider"`
	Model    string `gorm:"size:100;not null" json:"model"`
	BaseURL  string `gorm:"size:255;not null" json:"base_url"`

	// API Key 相关（json:"-"：永不返回给前端）
	APIKeyEnc     string `gorm:"type:text;not null" json:"-"`
	APIKeyPrefix  string `gorm:"size:10" json:"-"`
	APIKeySuffix  string `gorm:"size:10" json:"-"`

	IsEnabled bool `gorm:"default:true;not null;index" json:"is_enabled"`
	IsDefault bool `gorm:"default:false;not null;index" json:"is_default"`
}

// TableName 指定表名遵循 PRD 约定 ph_ 前缀。
func (l *LLMConfig) TableName() string {
	return "ph_llm_configs"
}