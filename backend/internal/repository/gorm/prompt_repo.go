// Package gorm - PromptRepo 的 GORM 实现。
package gorm

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/prompthub/prompthub/internal/model"
	"github.com/prompthub/prompthub/internal/repository"
)

type promptRepoGorm struct {
	db *gorm.DB
}

// NewPromptRepo 构造 GORM 版 PromptRepo。
func NewPromptRepo(db *gorm.DB) repository.PromptRepo {
	return &promptRepoGorm{db: db}
}

// Create 插入提示词。
func (r *promptRepoGorm) Create(ctx context.Context, p *model.Prompt) error {
	return r.db.WithContext(ctx).Create(p).Error
}

// GetByID 按主键 + 用户 ID 查询。
func (r *promptRepoGorm) GetByID(ctx context.Context, userID, id uint) (*model.Prompt, error) {
	var p model.Prompt
	err := r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, id).First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

// List 按过滤条件分页查询；返回 list 与 total。
func (r *promptRepoGorm) List(ctx context.Context, f repository.PromptListFilter) ([]*model.Prompt, int64, error) {
	page := f.Page
	if page <= 0 {
		page = 1
	}
	pageSize := f.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	qIn := f.QIn
	if qIn == "" {
		qIn = "all"
	}

	base := r.db.WithContext(ctx).Model(&model.Prompt{}).Where("user_id = ?", f.UserID)

	// 关键词
	if f.Keyword != "" {
		like := "%" + f.Keyword + "%"
		switch qIn {
		case "title":
			base = base.Where("title LIKE ?", like)
		default: // "all"
			base = base.Where("title LIKE ? OR description LIKE ? OR content LIKE ?", like, like, like)
		}
	}

	// 分类筛选
	if f.CategoryID != nil {
		base = base.Where("category_id = ?", *f.CategoryID)
	}

	// 收藏筛选
	if f.IsFavorite != nil {
		base = base.Where("is_favorite = ?", *f.IsFavorite)
	}

	// 计数
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	orderBy := "updated_at DESC"
	switch f.Sort {
	case "created_desc":
		orderBy = "created_at DESC"
	case "title_asc":
		orderBy = "title ASC"
	}

	var list []*model.Prompt
	err := base.
		Order(orderBy).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// Update 按字段更新。
func (r *promptRepoGorm) Update(ctx context.Context, userID, id uint, fields map[string]any) error {
	res := r.db.WithContext(ctx).
		Model(&model.Prompt{}).
		Where("user_id = ? AND id = ?", userID, id).
		Updates(fields)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete 软删。
func (r *promptRepoGorm) Delete(ctx context.Context, userID, id uint) error {
	res := r.db.WithContext(ctx).
		Where("user_id = ? AND id = ?", userID, id).
		Delete(&model.Prompt{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// BatchDelete 批量软删。
func (r *promptRepoGorm) BatchDelete(ctx context.Context, userID uint, ids []uint) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	res := r.db.WithContext(ctx).
		Where("user_id = ? AND id IN ?", userID, ids).
		Delete(&model.Prompt{})
	return res.RowsAffected, res.Error
}

// ToggleFavorite 切换收藏标记。
func (r *promptRepoGorm) ToggleFavorite(ctx context.Context, userID, id uint, isFavorite bool) error {
	res := r.db.WithContext(ctx).
		Model(&model.Prompt{}).
		Where("user_id = ? AND id = ?", userID, id).
		Update("is_favorite", isFavorite)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CountByCategory 统计当前用户下属于该分类的提示词数量。
func (r *promptRepoGorm) CountByCategory(ctx context.Context, userID, categoryID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Prompt{}).
		Where("user_id = ? AND category_id = ?", userID, categoryID).
		Count(&count).Error
	return count, err
}