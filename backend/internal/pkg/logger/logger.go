// Package logger 基于 uber-zap 的应用日志封装。
package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/prompthub/prompthub/internal/config"
)

// New 根据配置构造 *zap.Logger。
func New(cfg config.Logger) *zap.Logger {
	level := zap.NewAtomicLevel()
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encCfg := zap.NewProductionEncoderConfig()
		encCfg.TimeKey = "ts"
		encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		encoder = zapcore.NewJSONEncoder(encCfg)
	} else {
		encCfg := zap.NewDevelopmentEncoderConfig()
		encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encCfg)
	}

	var sink zapcore.WriteSyncer
	if cfg.OutputPath == "" || cfg.OutputPath == "stdout" {
		sink = zapcore.AddSync(os.Stdout)
	} else {
		f, err := os.OpenFile(cfg.OutputPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "日志文件打开失败，回退 stdout: %v\n", err)
			sink = zapcore.AddSync(os.Stdout)
		} else {
			sink = zapcore.AddSync(f)
		}
	}

	core := zapcore.NewCore(encoder, sink, level)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}
