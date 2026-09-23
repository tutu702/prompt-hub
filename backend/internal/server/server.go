// Package server 负责 HTTP 服务器的组装与生命周期管理。
package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/prompthub/prompthub/internal/config"
	"github.com/prompthub/prompthub/internal/handler"
	jwtpkg "github.com/prompthub/prompthub/internal/pkg/jwt"
	"github.com/prompthub/prompthub/internal/model"
	"github.com/prompthub/prompthub/internal/pkg/database"
	"github.com/prompthub/prompthub/internal/pkg/response"
	"github.com/prompthub/prompthub/internal/repository"
	gormrepo "github.com/prompthub/prompthub/internal/repository/gorm"
	"github.com/prompthub/prompthub/internal/router"
	"github.com/prompthub/prompthub/internal/service"
)

// HTTPServer 是后端 HTTP 服务器。
type HTTPServer struct {
	cfg    *config.Config
	logger *zap.Logger
	db     *gorm.DB
	router http.Handler
	srv    *http.Server
}

// New 构造 HTTP 服务器实例（不启动）。
func New(cfg *config.Config, log *zap.Logger) *HTTPServer {
	// 初始化数据库连接（失败不阻塞启动，便于本地仅验证 HTTP 时也能跑）
	db, dbErr := database.New(cfg.DB)
	if dbErr != nil {
		log.Warn("数据库连接失败，应用将以无数据库模式启动", zap.Error(dbErr))
		db = nil
	} else {
		log.Info("数据库连接成功",
			zap.String("driver", cfg.DB.Driver),
			zap.String("host", cfg.DB.Host),
			zap.Int("port", cfg.DB.Port),
		)
	}

	// AutoMigrate 业务表（骨架阶段：失败仅 warn；生产请替换为独立 migration 工具）
	if db != nil {
		if err := db.AutoMigrate(
			&model.User{},
			&model.Category{},
			&model.Prompt{},
		); err != nil {
			log.Warn("AutoMigrate 失败，登录注册/提示词相关接口将无法使用", zap.Error(err))
		} else {
			log.Info("AutoMigrate 完成：ph_users / ph_categories / ph_prompts")
		}
	}

	// 构造 JWT 管理器
	jwtMgr := jwtpkg.NewManager(cfg.JWT.Secret, cfg.JWT.Issuer, cfg.JWT.AccessExpire, cfg.JWT.RefreshExpire)
	log.Info("JWT 配置已加载",
		zap.Duration("access_expire", cfg.JWT.AccessExpire),
		zap.Duration("refresh_expire", cfg.JWT.RefreshExpire),
		zap.String("issuer", cfg.JWT.Issuer),
	)

	// 构造 repository / service / handler
	var (
		userRepo     repository.UserRepo
		categoryRepo repository.CategoryRepo
		promptRepo   repository.PromptRepo
	)
	if db != nil {
		userRepo = gormrepo.NewUserRepo(db)
		categoryRepo = gormrepo.NewCategoryRepo(db)
		promptRepo = gormrepo.NewPromptRepo(db)
	}

	authSvc := service.NewAuthService(userRepo, jwtMgr, log)
	categorySvc := service.NewCategoryService(categoryRepo, promptRepo.CountByCategory, log)
	promptSvc := service.NewPromptService(promptRepo, categoryRepo, log)

	healthH := handler.NewHealthHandler(cfg, db)
	authH := handler.NewAuthHandler(authSvc, log)
	categoryH := handler.NewCategoryHandler(categorySvc, log)
	promptH := handler.NewPromptHandler(promptSvc, log)

	engine := router.New(router.Options{
		Cfg:        cfg,
		Logger:     log,
		Health:     healthH,
		Auth:       authH,
		Category:   categoryH,
		Prompt:     promptH,
		JWTManager: jwtMgr,
	})

	srv := &http.Server{
		Addr:         cfg.Server.Host + ":" + itoa(cfg.Server.Port),
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	return &HTTPServer{
		cfg:    cfg,
		logger: log,
		db:     db,
		router: engine,
		srv:    srv,
	}
}

// Router 返回 HTTP 处理器，便于测试。
func (s *HTTPServer) Router() http.Handler {
	return s.router
}

// Run 启动服务，阻塞直到收到关闭信号。
func (s *HTTPServer) Run() error {
	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("HTTP 服务启动",
			zap.String("addr", s.srv.Addr),
			zap.String("mode", s.cfg.Server.Mode),
		)
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	if err := <-errCh; err != nil {
		return err
	}
	return nil
}

// Shutdown 优雅关闭。
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, s.cfg.Server.ShutdownTimeout)
	defer cancel()
	return s.srv.Shutdown(shutdownCtx)
}

// itoa 避免 strconv.Itoa 引入，本骨架足够简单，直接手写。
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// 保留 time import 已被使用避免未使用告警被某些 linter 误判。
var _ = time.Second

// 避免 response 包被未使用告警误判（部分 CI 会做全量引用检查）。
var _ = response.OK