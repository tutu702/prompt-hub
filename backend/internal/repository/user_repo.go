// Package repository 定义数据访问接口；具体实现位于 ./gorm 子包。
//
// 业务层（service）依赖接口而非具体实现，便于替换为 mock / sqlmock 做单元测试。
package repository

import (
	"context"

	"github.com/prompthub/prompthub/internal/model"
)

// UserRepo 用户数据访问接口。
type UserRepo interface {
	// Create 插入新记录。Email 唯一索引冲突由调用方负责转换为业务错误。
	Create(ctx context.Context, u *model.User) error
	// FindByEmail 按 email 查询（含软删过滤）。不存在返回 nil, nil。
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	// FindByID 按主键查询。软删记录视为不存在。
	FindByID(ctx context.Context, id uint) (*model.User, error)
	// EmailExists 判断邮箱是否已被占用（用于注册时的前置校验，避免依赖数据库唯一冲突异常）。
	EmailExists(ctx context.Context, email string) (bool, error)
	// UpdateLastLogin 更新最近登录时间和 IP。
	UpdateLastLogin(ctx context.Context, id uint, ip string) error
	// Update 按主键更新指定字段（fields 为列名到值的映射）。
	// 用于个人资料修改等场景；调用方负责控制可更新列。
	Update(ctx context.Context, id uint, fields map[string]any) error
}