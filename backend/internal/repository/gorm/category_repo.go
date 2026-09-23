// Package gorm - CategoryRepo 的 GORM 实现。
package gorm

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/prompthub/prompthub/internal/model"
	"github.com/prompthub/prompthub/internal/repository"
)

type categoryRepoGorm struct {
	db *gorm.DB
}

// NewCategoryRepo 构造 GORM 版 CategoryRepo。
func NewCategoryRepo(db *gorm.DB) repository.CategoryRepo {
	return &categoryRepoGorm{db: db}
}

// Create 插入分类。
func (r *categoryRepoGorm) Create(ctx context.Context, c *model.Category) error {
	return r.db.WithContext(ctx).Create(c).Error
}

// GetByID 按主键 + 用户 ID 查询。
func (r *categoryRepoGorm) GetByID(ctx context.Context, userID, id uint) (*model.Category, error) {
	var c model.Category
	err := r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, id).First(&c).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// List 按 sort ASC, created_at ASC 返回当前用户分类。
func (r *categoryRepoGorm) List(ctx context.Context, userID uint) ([]*model.Category, error) {
	var list []*model.Category
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("sort ASC, created_at ASC").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

// Update 更新指定字段。
func (r *categoryRepoGorm) Update(ctx context.Context, userID, id uint, fields map[string]any) error {
	res := r.db.WithContext(ctx).
		Model(&model.Category{}).
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
func (r *categoryRepoGorm) Delete(ctx context.Context, userID, id uint) error {
	res := r.db.WithContext(ctx).
		Where("user_id = ? AND id = ?", userID, id).
		Delete(&model.Category{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CountByID 校验存在性。
func (r *categoryRepoGorm) CountByID(ctx context.Context, userID, id uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Category{}).
		Where("user_id = ? AND id = ?", userID, id).
		Count(&count).Error
	return count, err
}