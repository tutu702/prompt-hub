// Package router 负责注册所有路由。
//
// 现阶段仅注册：
//   - GET /healthz
//   - GET /version
//   - 业务接口在后续迭代中按 PRD 顺序加入
package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/prompthub/prompthub/internal/config"
	"github.com/prompthub/prompthub/internal/handler"
	"github.com/prompthub/prompthub/internal/middleware"
)

// Options 路由注册需要的依赖。
type Options struct {
	Cfg      *config.Config
	Logger   *zap.Logger
	Health   *handler.HealthHandler
	// 后续业务控制器通过 Options 注入
	// UserHandler    *handler.UserHandler
	// PromptHandler  *handler.PromptHandler
}

// New 构建 *gin.Engine。
func New(opt Options) *gin.Engine {
	gin.SetMode(opt.Cfg.Server.Mode)

	r := gin.New()
	r.Use(
		middleware.Recovery(opt.Logger),
		middleware.AccessLog(opt.Logger),
		middleware.CORS(),
	)

	// 基础探测
	r.GET("/healthz", opt.Health.Healthz)

	// API 路由组
	api := r.Group("/api/v1")
	{
		api.GET("/version", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"app":     opt.Cfg.App.Name,
				"version": opt.Cfg.App.Version,
				"env":     opt.Cfg.App.Env,
			})
		})

		// TODO: 业务路由按 PRD 章节 8 添加
		// auth := api.Group("/auth")
		// prompts := api.Group("/prompts")
		// categories := api.Group("/categories")
		// llm := api.Group("/llm")
	}

	// 兜底 404
	r.NoRoute(handler.NotFound)
	r.NoMethod(handler.NotFound)

	return r
}
