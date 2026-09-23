// Package model - Category 分类表（ph_categories）。
//
// 用于把用户的提示词按"编程 / 写作 / 研究"等自定义分类归档。
// 同一用户下分类名称不可重复（软删场景下唯一性由部分唯一索引保证）。
package model

import (
	"time"

	"gorm.io/gorm"
)

// Category 分类模型。
//
// 字段语义：
//   - UserID：所属用户；隔离数据。
//   - Name：分类名；同一用户下唯一（软删后允许同名新建）。
//   - Color：HEX 颜色（#RRGGBB），用于卡片标签；本期前端使用 Element Plus 默认色。
//   - Sort：升序展示；默认 0。
type Category struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	Name      string         `gorm:"size:20;not null" json:"name"`
	Color     string         `gorm:"size:7;default:''" json:"color"`
	Sort      int            `gorm:"default:0;not null;index" json:"sort"`
}

// TableName 指定表名遵循 PRD 约定 ph_ 前缀。
func (c *Category) TableName() string {
	return "ph_categories"
}