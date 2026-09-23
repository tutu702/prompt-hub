// Package database 提供基于 GORM 的数据库连接初始化能力（不创建业务表）。
//
// 默认驱动由配置 db.driver 决定：postgres 或 mysql。
// 初始化时仅做连通性测试，不进行 AutoMigrate，避免骨架阶段产生业务表结构。
package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/prompthub/prompthub/internal/config"
)

// New 根据配置构造 *gorm.DB。
//
// 注意：本骨架不调用 AutoMigrate，业务模型在后续迭代中按需迁移。
func New(cfg config.Database) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		Logger:                                   newGormLogger(cfg.LogLevel),
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: true,
		NowFunc:                                  func() time.Time { return time.Now().UTC() },
	}

	switch cfg.Driver {
	case "postgres":
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
		)
		return gorm.Open(postgres.Open(dsn), gormCfg)
	case "mysql":
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.Charset,
		)
		return gorm.Open(mysql.Open(dsn), gormCfg)
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %s", cfg.Driver)
	}
}

// Ping 校验数据库连通性。
func Ping(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// WithTx 在事务中运行 fn。fn 返回 nil → Commit；返回 err → Rollback。
// fn 的返回值会被原样作为 WithTx 的返回值透传给调用方。
//
// 使用示例：
//
//	err := database.WithTx(ctx, db, func(tx *gorm.DB) error {
//	    if err := tx.Create(&a).Error; err != nil { return err }
//	    if err := tx.Create(&b).Error; err != nil { return err }
//	    return nil
//	})
func WithTx(ctx context.Context, db *gorm.DB, fn func(tx *gorm.DB) error) (err error) {
	if db == nil {
		return gorm.ErrInvalidDB
	}
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback().Error
			panic(p)
		}
		if err != nil {
			if rbErr := tx.Rollback().Error; rbErr != nil {
				err = fmt.Errorf("%w (rollback err: %v)", err, rbErr)
			}
			return
		}
		if cmErr := tx.Commit().Error; cmErr != nil {
			err = cmErr
		}
	}()
	err = fn(tx)
	return
}

// stdPrintfWriter 将 os.Stdout 适配到 gorm logger.Writer 接口（仅 Printf）。
type stdPrintfWriter struct{}

func (stdPrintfWriter) Printf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, format, args...)
}

func newGormLogger(level string) gormlogger.Interface {
	writer := stdPrintfWriter{}
	cfg := gormlogger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  gormlogger.Warn,
		IgnoreRecordNotFoundError: true,
		Colorful:                  true,
	}
	switch level {
	case "silent":
		cfg.LogLevel = gormlogger.Silent
	case "error":
		cfg.LogLevel = gormlogger.Error
	case "info":
		cfg.LogLevel = gormlogger.Info
	default:
		cfg.LogLevel = gormlogger.Warn
	}
	return gormlogger.New(writer, cfg)
}
