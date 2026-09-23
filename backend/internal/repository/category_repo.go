// Package repository - Category 数据访问接口。
//
// 所有方法强制 WHERE user_id=?；不存在时返回 (nil, nil)（service 层判空后转 404）。
package repository

import (
	"context"

	"github.com/prompthub/prompthub/internal/model"
)

// CategoryRepo 分类数据访问接口。
type CategoryRepo interface {
	// Create 插入分类。
	Create(ctx context.Context, c *model.Category) error
	// GetByID 按主键 + 用户 ID 查询；软删记录视为不存在。
	GetByID(ctx context.Context, userID, id uint) (*model.Category, error)
	// List 列出当前用户的所有分类（按 sort 升序、created_at 升序）。
	List(ctx context.Context, userID uint) ([]*model.Category, error)
	// Update 按 userID + id 更新指定字段（fields 为列名到值的映射）。
	Update(ctx context.Context, userID, id uint, fields map[string]any) error
	// Delete 按 userID + id 软删。
	Delete(ctx context.Context, userID, id uint) error
	// CountByID 判断指定 userID 下该 id 是否存在（用于删除前的存在性校验）。
	CountByID(ctx context.Context, userID, id uint) (int64, error)
}