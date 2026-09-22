// Package main 是 PromptHub 后端的启动入口。
//
// 使用 GIN_WEB 企业级落地标准：
//   cmd/server          启动入口
//   internal/config     Viper 配置加载
//   internal/server     HTTP 服务器装配
//   internal/router     路由注册
//   internal/handler    HTTP 处理层
//   internal/middleware 中间件（日志、恢复、跨域）
//   internal/service    业务编排（本骨架不实现）
//   internal/repository 数据访问（本骨架不实现）
//   internal/model      数据模型 / GORM 模型（本骨架不实现）
//   internal/pkg        数据库、响应、日志等基础设施
//   pkg                 通用工具与错误码
//
// 现阶段仅启动 HTTP 服务并暴露 /healthz 一个健康检查端点。
package main

import (
	"fmt"
	"os"

	"github.com/prompthub/prompthub/internal/config"
	"github.com/prompthub/prompthub/internal/pkg/logger"
	"github.com/prompthub/prompthub/internal/server"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化日志
	log := logger.New(cfg.Log)
	defer func() { _ = log.Sync() }()

	// 3. 启动 HTTP 服务（阻塞）
	srv := server.New(cfg, log)
	if err := srv.Run(); err != nil {
		log.Sugar().Fatalf("HTTP 服务异常退出: %v", err)
	}
}
