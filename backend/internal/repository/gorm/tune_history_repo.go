// Package gorm - TuneHistoryRepo 的 GORM 实现。
package gorm

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/prompthub/prompthub/internal/model"
	"github.com/prompthub/prompthub/internal/repository"
)

type tuneHistoryRepoGorm struct {
	db *gorm.DB
}

// NewTuneHistoryRepo 构造 GORM 版 TuneHistoryRepo。
func NewTuneHistoryRepo(db *gorm.DB) repository.TuneHistoryRepo {
	return &tuneHistoryRepoGorm{db: db}
}

// Create 插入一条历史记录。
func (r *tuneHistoryRepoGorm) Create(ctx context.Context, h *model.TuneHistory) error {
	return r.db.WithContext(ctx).Create(h).Error
}

// GetByID 按主键 + userID 查询。
func (r *tuneHistoryRepoGorm) GetByID(ctx context.Context, userID, id uint) (*model.TuneHistory, error) {
	var h model.TuneHistory
	err := r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, id).First(&h).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &h, nil
}

// List 按 prompt 列出某 user 的历史（version_no DESC 分页）。
func (r *tuneHistoryRepoGorm) List(ctx context.Context, f repository.TuneHistoryListFilter) ([]*model.TuneHistory, int64, error) {
	page, size := f.Page, f.PageSize
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}

	base := r.db.WithContext(ctx).Model(&model.TuneHistory{}).
		Where("user_id = ? AND prompt_id = ?", f.UserID, f.PromptID)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.TuneHistory
	err := base.Order("version_no DESC").
		Offset((page - 1) * size).Limit(size).
		Find(&list).Error
	return list, total, err
}

// MaxVersionNo 取某 prompt 下 user 的最大 version_no（0 表示无）。
func (r *tuneHistoryRepoGorm) MaxVersionNo(ctx context.Context, userID, promptID uint) (int, error) {
	var maxV int
	err := r.db.WithContext(ctx).Model(&model.TuneHistory{}).
		Where("user_id = ? AND prompt_id = ?", userID, promptID).
		Select("COALESCE(MAX(version_no), 0)").
		Scan(&maxV).Error
	return maxV, err
}

// MarkApplied 设置 is_applied=true。
func (r *tuneHistoryRepoGorm) MarkApplied(ctx context.Context, userID, id uint) error {
	res := r.db.WithContext(ctx).Model(&model.TuneHistory{}).
		Where("user_id = ? AND id = ?", userID, id).
		Update("is_applied", true)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}