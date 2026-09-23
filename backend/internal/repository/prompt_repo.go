// Package repository - Prompt 数据访问接口。
//
// 所有读写强制 WHERE user_id=?；不存在时返回 (nil, nil)。
// 软删过滤由 GORM DeletedAt 自动处理。
package repository

import (
	"context"

	"github.com/prompthub/prompthub/internal/model"
)

// PromptListFilter 列表查询过滤条件。
type PromptListFilter struct {
	UserID     uint
	Keyword    string // 关键词；空串表示不过滤
	QIn        string // "title" 或 "all"，空字符串默认 "all"
	CategoryID *uint  // nil 表示全部
	IsFavorite *bool  // nil 表示全部
	Sort       string // "updated_desc" / "created_desc" / "title_asc"
	Page       int    // 1-based；<=0 视为 1
	PageSize   int    // 默认 20，上限 100
}

// PromptRepo 提示词数据访问接口。
type PromptRepo interface {
	// Create 插入新提示词。
	Create(ctx context.Context, p *model.Prompt) error
	// GetByID 按主键 + 用户 ID 查询。
	GetByID(ctx context.Context, userID, id uint) (*model.Prompt, error)
	// List 按过滤条件分页查询，返回 list 与 total。
	List(ctx context.Context, filter PromptListFilter) ([]*model.Prompt, int64, error)
	// Update 按 userID + id 更新指定字段。
	Update(ctx context.Context, userID, id uint, fields map[string]any) error
	// Delete 按 userID + id 软删。
	Delete(ctx context.Context, userID, id uint) error
	// BatchDelete 批量软删；仅软删属于 userID 的记录，越权 ID 不会影响他人数据。
	BatchDelete(ctx context.Context, userID uint, ids []uint) (int64, error)
	// ToggleFavorite 切换收藏标记（is_favorite 由调用方决定最终值）。
	ToggleFavorite(ctx context.Context, userID, id uint, isFavorite bool) error
	// CountByCategory 统计指定 userID 下属于该分类的提示词数量。
	CountByCategory(ctx context.Context, userID, categoryID uint) (int64, error)
}