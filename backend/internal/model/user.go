// Package model 定义 GORM 持久化模型。
//
// 当前骨架已就绪：ph_users 表用于登录/注册场景，由 server 启动时 AutoMigrate 创建。
// 其它业务模型（ph_categories / ph_prompts 等）按 PRD §七 后续迭代添加。
package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户表模型。
//
// 字段语义：
//   - Email：登录邮箱；同一库内唯一（软删场景下唯一索引见 PRD §七 7.1）。
//   - PasswordHash：bcrypt 哈希；不存明文。
//   - Nickname：用户昵称；为空时取 email 本地部分。
//   - TokenVersion：用于未来"修改密码/下线所有设备"等场景 bump 掉旧 refresh token；本期签发时写入但不校验。
//   - Status：1=active 0=disabled；本期不暴露禁用入口。
//   - LastLoginAt / LastLoginIP：登录审计使用。
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Email        string         `gorm:"size:191;uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"`
	Nickname     string         `gorm:"size:64" json:"nickname"`
	TokenVersion int            `gorm:"default:1;not null" json:"-"`
	Status       int            `gorm:"default:1;not null" json:"-"`
	LastLoginAt  *time.Time     `json:"last_login_at,omitempty"`
	LastLoginIP  string         `gorm:"size:64" json:"-"`
}

// TableName 指定表名遵循 PRD 约定 ph_ 前缀。
func (u *User) TableName() string {
	return "ph_users"
}