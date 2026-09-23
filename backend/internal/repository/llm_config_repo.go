// Package repository - LLMConfigRepo 大模型配置数据访问接口。
package repository

import (
	"context"

	"github.com/prompthub/prompthub/internal/model"
)

// LLMConfigRepo 大模型配置数据访问接口。
type LLMConfigRepo interface {
	// Create 插入一条配置。
	Create(ctx context.Context, c *model.LLMConfig) error
	// GetByID 按主键 + userID 查询；越权返 (nil, nil)。
	GetByID(ctx context.Context, userID, id uint) (*model.LLMConfig, error)
	// List 列出某 user 的全部配置。
	List(ctx context.Context, userID uint) ([]*model.LLMConfig, error)
	// ListEnabled 列出某 user 的启用配置（给 AI 创建 / 调优挑选用）。
	ListEnabled(ctx context.Context, userID uint) ([]*model.LLMConfig, error)
	// GetDefault 取某 user 的默认配置；不存在返 (nil, nil)。
	GetDefault(ctx context.Context, userID uint) (*model.LLMConfig, error)
	// Update 按 userID + id 更新指定字段。
	Update(ctx context.Context, userID, id uint, fields map[string]any) error
	// Delete 按 userID + id 软删。
	Delete(ctx context.Context, userID, id uint) error
	// ClearDefault 把某 user 全部配置的 is_default 清掉（在事务里调用）。
	ClearDefault(ctx context.Context, userID uint) error
	// CountByID 校验存在性（软删返 0）。
	CountByID(ctx context.Context, userID, id uint) (int64, error)
	// WithTx 事务封装（用于 SetDefault 等多步操作）。
	WithTx(ctx context.Context, fn func(repo LLMConfigRepo) (any, error)) (any, error)
}