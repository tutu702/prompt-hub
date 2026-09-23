// Package gorm - LLMConfigRepo 的 GORM 实现。
package gorm

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/prompthub/prompthub/internal/model"
	"github.com/prompthub/prompthub/internal/repository"
)

type llmConfigRepoGorm struct {
	db *gorm.DB
}

// wrapTx 把 gorm tx 包成 LLMConfigRepo 走同一组方法（保持 userID 强校验等逻辑统一）。
func wrapTx(tx *gorm.DB) repository.LLMConfigRepo {
	return &llmConfigRepoGorm{db: tx}
}

// NewLLMConfigRepo 构造 GORM 版 LLMConfigRepo。
func NewLLMConfigRepo(db *gorm.DB) repository.LLMConfigRepo {
	return &llmConfigRepoGorm{db: db}
}

func (r *llmConfigRepoGorm) Create(ctx context.Context, c *model.LLMConfig) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *llmConfigRepoGorm) GetByID(ctx context.Context, userID, id uint) (*model.LLMConfig, error) {
	var c model.LLMConfig
	err := r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, id).First(&c).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *llmConfigRepoGorm) List(ctx context.Context, userID uint) ([]*model.LLMConfig, error) {
	var list []*model.LLMConfig
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_default DESC, is_enabled DESC, id ASC").
		Find(&list).Error
	return list, err
}

func (r *llmConfigRepoGorm) ListEnabled(ctx context.Context, userID uint) ([]*model.LLMConfig, error) {
	var list []*model.LLMConfig
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_enabled = ?", userID, true).
		Order("is_default DESC, id ASC").
		Find(&list).Error
	return list, err
}

func (r *llmConfigRepoGorm) GetDefault(ctx context.Context, userID uint) (*model.LLMConfig, error) {
	var c model.LLMConfig
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_default = ?", userID, true).
		First(&c).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *llmConfigRepoGorm) Update(ctx context.Context, userID, id uint, fields map[string]any) error {
	res := r.db.WithContext(ctx).
		Model(&model.LLMConfig{}).
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

func (r *llmConfigRepoGorm) Delete(ctx context.Context, userID, id uint) error {
	res := r.db.WithContext(ctx).
		Where("user_id = ? AND id = ?", userID, id).
		Delete(&model.LLMConfig{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *llmConfigRepoGorm) ClearDefault(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).
		Model(&model.LLMConfig{}).
		Where("user_id = ?", userID).
		Update("is_default", false).Error
}

func (r *llmConfigRepoGorm) CountByID(ctx context.Context, userID, id uint) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.LLMConfig{}).
		Where("user_id = ? AND id = ?", userID, id).
		Count(&n).Error
	return n, err
}

func (r *llmConfigRepoGorm) WithTx(ctx context.Context, fn func(repo repository.LLMConfigRepo) (any, error)) (any, error) {
	var result any
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var fnErr error
		result, fnErr = fn(wrapTx(tx))
		return fnErr
	})
	return result, err
}