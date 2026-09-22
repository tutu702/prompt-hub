// Package gorm 提供 UserRepo 的 GORM 实现。
//
// 所有查询均带 WithContext(ctx) 以透传请求级超时与链路信息；软删由 gorm.DeletedAt 自动处理。
package gorm

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/prompthub/prompthub/internal/model"
	"github.com/prompthub/prompthub/internal/repository"
)

// userRepoGorm 是 UserRepo 的 GORM 实现。
type userRepoGorm struct {
	db *gorm.DB
}

// NewUserRepo 构造 GORM 版 UserRepo。
func NewUserRepo(db *gorm.DB) repository.UserRepo {
	return &userRepoGorm{db: db}
}

// Create 插入新用户。
func (r *userRepoGorm) Create(ctx context.Context, u *model.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

// FindByEmail 按 email 查询。GORM 默认会过滤 deleted_at IS NOT NULL 的记录。
func (r *userRepoGorm) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// FindByID 按主键查询。
func (r *userRepoGorm) FindByID(ctx context.Context, id uint) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).First(&u, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// EmailExists 判断邮箱是否已被占用（不计软删记录）。
func (r *userRepoGorm) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateLastLogin 写入最近登录时间和 IP（best-effort，错误由调用方记日志）。
func (r *userRepoGorm) UpdateLastLogin(ctx context.Context, id uint, ip string) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_login_at": now,
			"last_login_ip": ip,
		}).Error
}