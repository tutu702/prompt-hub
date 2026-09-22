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
	"github.com/prompthub/prompthub/internal/pkg/database"
	"github.com/prompthub/prompthub/internal/router"
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

	healthH := handler.NewHealthHandler(cfg, db)
	engine := router.New(router.Options{
		Cfg:    cfg,
		Logger: log,
		Health: healthH,
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
