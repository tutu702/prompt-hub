// Package server 负责 HTTP 服务器的组装与生命周期管理。
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/prompthub/prompthub/internal/config"
	"github.com/prompthub/prompthub/internal/handler"
	jwtpkg "github.com/prompthub/prompthub/internal/pkg/jwt"
	cryptopkg "github.com/prompthub/prompthub/internal/pkg/crypto"
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
			&model.TuneHistory{},
			&model.LLMConfig{},
		); err != nil {
			log.Warn("AutoMigrate 失败，登录注册/提示词相关接口将无法使用", zap.Error(err))
		} else {
			log.Info("AutoMigrate 完成：ph_users / ph_categories / ph_prompts / ph_tune_histories / ph_llm_configs")
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
		userRepo           repository.UserRepo
		categoryRepo       repository.CategoryRepo
		promptRepo         repository.PromptRepo
		tuneHistoryRepo    repository.TuneHistoryRepo
	)
	if db != nil {
		userRepo = gormrepo.NewUserRepo(db)
		categoryRepo = gormrepo.NewCategoryRepo(db)
		promptRepo = gormrepo.NewPromptRepo(db)
		tuneHistoryRepo = gormrepo.NewTuneHistoryRepo(db)
	}

	authSvc := service.NewAuthService(userRepo, jwtMgr, log)
	categorySvc := service.NewCategoryService(categoryRepo, promptRepo.CountByCategory, log)
	promptSvc := service.NewPromptService(promptRepo, categoryRepo, log)
	tuneHistorySvc := service.NewTuneHistoryService(tuneHistoryRepo, promptRepo, log)

	healthH := handler.NewHealthHandler(cfg, db)
	authH := handler.NewAuthHandler(authSvc, log)
	categoryH := handler.NewCategoryHandler(categorySvc, log)
	promptH := handler.NewPromptHandler(promptSvc, log)
	tuneHistoryH := handler.NewTuneHistoryHandler(tuneHistorySvc, log)

	// P08 LLM 加密：API Key 加密密钥。
	//
	// 解析顺序（env > file > auto-generated）：
	//   1. PROMPTHUB_LLM_AES_KEY 环境变量直接给 hex（生产 / 集群显式控制，便于多实例共享 + 轮换）
	//   2. 持久文件（缺省 ./data/llm_aes.key）—— 首次启动自动生成 32 字节随机 key 落盘（0600），
	//      之后重启沿用同一把。这样开箱即用，且历史加密数据不会"丢钥匙"。
	//   3. 都没拿到 —— 生成内存 key 并 WARN；重启即失效（用户必须在生产显式配置）。
	aesSvc := loadLLMAESKey(cfg, log)

	llmConfigRepo := gormrepo.NewLLMConfigRepo(db)
	llmConfigSvc := service.NewLLMConfigService(llmConfigRepo, aesSvc, log)
	llmConfigH := handler.NewLLMConfigHandler(llmConfigSvc, log)

	// P07 调优服务（依赖 PromptRepo + LLMConfigRepo + TuneHistoryRepo + AES）
	llmTuneSvc := service.NewLLMTuneService(promptRepo, llmConfigRepo, tuneHistoryRepo, aesSvc, log)
	llmTuneH := handler.NewLLMTuneHandler(llmTuneSvc, log)

	// P04 / F08 AI 生成服务
	llmGenerateSvc := service.NewLLMGenerateService(llmConfigRepo, aesSvc, log)
	llmGenerateH := handler.NewLLMGenerateHandler(llmGenerateSvc, log)

	engine := router.New(router.Options{
		Cfg:         cfg,
		Logger:      log,
		Health:      healthH,
		Auth:        authH,
		Category:    categoryH,
		Prompt:      promptH,
		TuneHistory: tuneHistoryH,
		LLMConfig:   llmConfigH,
		LLMTune:     llmTuneH,
		LLMGenerate: llmGenerateH,
		JWTManager:  jwtMgr,
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

// loadLLMAESKey 按 env > file > auto-generated 顺序加载 / 生成 LLM AES 主密钥。
//
//   - env     PROMPTHUB_LLM_AES_KEY：64 字符 hex，最优先（生产显式控制）
//   - file    cfg.LLM.KeyFile（默认 ./data/llm_aes.key）：缺则生成 32 字节随机 key 落盘（0600），
//     存在则直接读取；读到非法 hex 视为数据损坏，fail-fast 不覆盖（防止静默丢钥匙）
//   - in-mem  都没有：生成随机 key 仅内存持有 + WARN；下次重启即失效
//
// 返回 nil 表示彻底失败（加密功能不可用，所有 /llm-configs / /llm/tune / /llm/generate 接口将 503）。
func loadLLMAESKey(cfg *config.Config, log *zap.Logger) *cryptopkg.AES {
	// 1) 优先级最高：环境变量
	if cfg.LLM.AESKey != "" {
		a, err := cryptopkg.NewFromHex(cfg.LLM.AESKey)
		if err != nil {
			log.Error("LLM AES key 解析失败（环境变量 PROMPTHUB_LLM_AES_KEY）",
				zap.String("source", "env"),
				zap.Error(err),
			)
			return nil
		}
		log.Info("LLM AES key 已加载（来源：环境变量 PROMPTHUB_LLM_AES_KEY）",
			zap.String("source", "env"),
		)
		return a
	}

	// 2) 持久文件
	keyFile := strings.TrimSpace(cfg.LLM.KeyFile)
	if keyFile != "" {
		// 先尝试读；缺失则走"生成并落盘"分支
		if data, err := os.ReadFile(keyFile); err == nil {
			hexKey := strings.TrimSpace(string(data))
			if a, err := cryptopkg.NewFromHex(hexKey); err == nil {
				log.Info("LLM AES key 已加载（来源：持久文件）",
					zap.String("source", "file"),
					zap.String("path", keyFile),
				)
				return a
			} else {
				// 关键安全策略：读到非法 hex 视为数据损坏，绝不静默覆盖。
				// 否则旧加密数据会全部变成「密文乱码」，用户毫无察觉。
				log.Error("LLM AES key 文件内容非法（不是 64 字符 hex）。为防止覆盖现有密文，已停止启动。请检查文件或显式配置 PROMPTHUB_LLM_AES_KEY",
					zap.String("source", "file"),
					zap.String("path", keyFile),
					zap.Error(err),
				)
				return nil
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			log.Error("LLM AES key 文件读取失败",
				zap.String("source", "file"),
				zap.String("path", keyFile),
				zap.Error(err),
			)
			return nil
		}

		// 文件不存在 → 生成并落盘
		newKey, err := cryptopkg.GenerateKey()
		if err != nil {
			log.Error("LLM AES key 生成失败", zap.Error(err))
			return nil
		}
		if err := writeAESKeyFile(keyFile, newKey); err != nil {
			// 落盘失败（权限 / 只读 FS 等）。仍然把生成的 key 用上，但下次重启将失效。
			// 一旦 LLMConfigs 表里写了密文，第二次启动就解不开，所以必须 WARN。
			log.Warn("LLM AES key 已生成但落盘失败（下次重启将无法解密历史 API Key）；生产部署请配置 PROMPTHUB_LLM_AES_KEY 或修复文件路径权限",
				zap.String("source", "generated(in-memory only)"),
				zap.String("path", keyFile),
				zap.Error(err),
			)
			a, _ := cryptopkg.NewFromHex(newKey)
			return a
		}
		a, _ := cryptopkg.NewFromHex(newKey)
		log.Info("LLM AES key 已自动生成并落盘（首次启动）。生产 / 多实例部署请改用 PROMPTHUB_LLM_AES_KEY 以保证实例间 key 一致",
			zap.String("source", "generated"),
			zap.String("path", keyFile),
		)
		return a
	}

	// 3) 都没配置：纯内存 key + WARN
	newKey, err := cryptopkg.GenerateKey()
	if err != nil {
		log.Error("LLM AES key 生成失败", zap.Error(err))
		return nil
	}
	log.Warn("未配置 LLM AES key（既无 PROMPTHUB_LLM_AES_KEY 也无 key_file）。已生成临时 key，重启即失效；生产部署必须显式配置",
		zap.String("source", "generated(in-memory, ephemeral)"),
	)
	a, _ := cryptopkg.NewFromHex(newKey)
	return a
}

// writeAESKeyFile 把生成的 key 落盘，自动 mkdir -p 父目录，权限 0600。
func writeAESKeyFile(path, hexKey string) error {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return fmt.Errorf("创建 key 文件目录失败: %w", err)
		}
	}
	// 0600：仅 owner 可读写，避免其他人读到 LLM API Key 主密钥
	if err := os.WriteFile(path, []byte(hexKey+"\n"), 0o600); err != nil {
		return err
	}
	return nil
}