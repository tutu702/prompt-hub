// Package router 负责注册所有路由。
//
// 当前已注册：
//   - GET /healthz
//   - GET /api/v1/version
//   - /api/v1/auth/{register,login,refresh}（公开）
//   - /api/v1/auth/logout（受 JWT 中间件保护）
//   - /api/v1/prompts/*（受 JWT 保护）
//   - /api/v1/categories/*（受 JWT 保护）
//
// 其它业务接口按 PRD 章节 8 在后续迭代加入。
package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	jwtpkg "github.com/prompthub/prompthub/internal/pkg/jwt"

	"github.com/prompthub/prompthub/internal/config"
	"github.com/prompthub/prompthub/internal/handler"
	"github.com/prompthub/prompthub/internal/middleware"
)

// Options 路由注册需要的依赖。
type Options struct {
	Cfg        *config.Config
	Logger     *zap.Logger
	Health     *handler.HealthHandler
	Auth       *handler.AuthHandler
	Category   *handler.CategoryHandler
	Prompt     *handler.PromptHandler
	JWTManager *jwtpkg.Manager
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
		// 版本信息
		api.GET("/version", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"app":     opt.Cfg.App.Name,
				"version": opt.Cfg.App.Version,
				"env":     opt.Cfg.App.Env,
			})
		})

		// ---- 鉴权：公开 ----
		auth := api.Group("/auth")
		{
			auth.POST("/register", opt.Auth.Register)
			auth.POST("/login", opt.Auth.Login)
			auth.POST("/refresh", opt.Auth.Refresh)
		}

		// ---- 受保护 ----
		protected := api.Group("")
		protected.Use(middleware.JWT(opt.JWTManager, opt.Logger))
		{
			protected.POST("/auth/logout", opt.Auth.Logout)

			// 提示词
			protected.GET("/prompts", opt.Prompt.List)
			protected.POST("/prompts", opt.Prompt.Create)
			protected.GET("/prompts/:id", opt.Prompt.Get)
			protected.PUT("/prompts/:id", opt.Prompt.Update)
			protected.DELETE("/prompts/:id", opt.Prompt.Delete)
			protected.POST("/prompts/batch-delete", opt.Prompt.BatchDelete)
			protected.POST("/prompts/:id/favorite", opt.Prompt.ToggleFavorite)

			// 分类
			protected.GET("/categories", opt.Category.List)
			protected.POST("/categories", opt.Category.Create)
			protected.PUT("/categories/:id", opt.Category.Update)
			protected.DELETE("/categories/:id", opt.Category.Delete)
		}
	}

	// 兜底 404
	r.NoRoute(handler.NotFound)
	r.NoMethod(handler.NotFound)

	return r
}