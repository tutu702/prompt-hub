// Package repository - TuneHistoryRepo 调优历史数据访问接口。
package repository

import (
	"context"

	"github.com/prompthub/prompthub/internal/model"
)

// TuneHistoryListFilter 列表查询过滤条件。
type TuneHistoryListFilter struct {
	UserID   uint
	PromptID uint
	Page     int
	PageSize int
}

// TuneHistoryRepo 调优历史数据访问接口。
type TuneHistoryRepo interface {
	// Create 插入一条历史记录。
	Create(ctx context.Context, h *model.TuneHistory) error
	// GetByID 按主键 + userID 查询；越权返回 (nil, nil)。
	GetByID(ctx context.Context, userID, id uint) (*model.TuneHistory, error)
	// List 按 prompt 列出某 user 的历史（version_no DESC）。
	List(ctx context.Context, f TuneHistoryListFilter) ([]*model.TuneHistory, int64, error)
	// MaxVersionNo 取某 prompt 下 user 的最大 version_no（0 表示无）。
	MaxVersionNo(ctx context.Context, userID, promptID uint) (int, error)
	// MarkApplied 设置 is_applied=true（幂等）。
	MarkApplied(ctx context.Context, userID, id uint) error
}