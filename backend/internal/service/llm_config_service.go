// Package service - LLMConfigService 大模型配置业务编排（PRD P08）。
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/prompthub/prompthub/internal/model"
	"github.com/prompthub/prompthub/internal/pkg/crypto"
	"github.com/prompthub/prompthub/internal/repository"
	"github.com/prompthub/prompthub/pkg/errcode"
)

// LLMConfigService 大模型配置服务。
type LLMConfigService struct {
	repo repository.LLMConfigRepo
	aes  *crypto.AES // nil 表示 AES 未启用 → 所有读写/测试连接返回 503
	http *http.Client
	log  *zap.Logger
}

// NewLLMConfigService 构造 LLMConfigService。
func NewLLMConfigService(
	repo repository.LLMConfigRepo,
	aes *crypto.AES,
	log *zap.Logger,
) *LLMConfigService {
	return &LLMConfigService{
		repo: repo,
		aes:  aes,
		http: &http.Client{Timeout: 10 * time.Second},
		log: log,
	}
}

// allowedProviders 7 种供应商（PRD §五 P08 字段枚举）。
var allowedProviders = map[string]struct{}{
	"openai":    {},
	"anthropic": {},
	"deepseek":  {},
	"qwen":      {},
	"zhipu":     {},
	"moonshot":  {},
	"custom":    {},
}

// LLMConfigInput 创建/更新入参（APIKey 为空串表示更新时不改）。
type LLMConfigInput struct {
	Name      string
	Provider  string
	Model     string
	BaseURL   string
	APIKey    string
	IsEnabled *bool
	IsDefault *bool
}

// TestResult 测试连接结果（service 永远返回它，handler 转 DTO）。
type TestResult struct {
	OK           bool
	LatencyMs    int64
	ErrorCode    string
	ErrorMessage string
}

const (
	testCodeAuthFailed     = "AUTH_FAILED"
	testCodeTimeout        = "TIMEOUT"
	testCodeModelNotFound  = "MODEL_NOT_FOUND"
	testCodeQuotaExhausted = "QUOTA_EXHAUSTED"
	testCodeNetworkError   = "NETWORK_ERROR"
)

// Create 创建一条配置。
func (s *LLMConfigService) Create(
	ctx context.Context,
	userID uint,
	in LLMConfigInput,
) (*model.LLMConfig, error) {
	if err := s.validateInput(in, true); err != nil {
		return nil, err
	}

	// 同名预检
	existing, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, c := range existing {
		if c.Name == in.Name {
			return nil, newAppError(errcode.CodeConflict, "已存在同名配置")
		}
	}

	enc, prefix, suffix, err := s.encryptKey(in.APIKey)
	if err != nil {
		return nil, err
	}

	isEnabled := true
	if in.IsEnabled != nil {
		isEnabled = *in.IsEnabled
	}
	isDefault := false
	if in.IsDefault != nil {
		isDefault = *in.IsDefault
	}
	if isDefault && !isEnabled {
		return nil, newAppError(errcode.CodeBadRequest, "默认模型必须同时启用")
	}

	cfg := &model.LLMConfig{
		UserID:        userID,
		Name:          in.Name,
		Provider:      in.Provider,
		Model:         in.Model,
		BaseURL:       in.BaseURL,
		APIKeyEnc:     enc,
		APIKeyPrefix:  prefix,
		APIKeySuffix:  suffix,
		IsEnabled:     isEnabled,
		IsDefault:     isDefault,
	}

	if isDefault {
		// 事务：ClearDefault + Create
		_, err := s.repo.WithTx(ctx, func(r repository.LLMConfigRepo) (any, error) {
			if err := r.ClearDefault(ctx, userID); err != nil {
				return nil, err
			}
			return nil, r.Create(ctx, cfg)
		})
		if err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.Create(ctx, cfg); err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

// Update 更新（字段增量；APIKey 为空表示不改）。
func (s *LLMConfigService) Update(
	ctx context.Context,
	userID, id uint,
	in LLMConfigInput,
) (*model.LLMConfig, error) {
	if in.Name == "" && in.Provider == "" && in.Model == "" && in.BaseURL == "" &&
		in.APIKey == "" && in.IsEnabled == nil && in.IsDefault == nil {
		return nil, newAppError(errcode.CodeBadRequest, "没有需要更新的内容")
	}

	existing, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, newAppError(errcode.CodeNotFound, "大模型配置不存在")
	}

	fields := map[string]any{}

	if in.Name != "" {
		if len(in.Name) > 20 {
			return nil, newAppError(errcode.CodeBadRequest, "名称长度需为 1-20")
		}
		// 同名校验
		all, err := s.repo.List(ctx, userID)
		if err != nil {
			return nil, err
		}
		for _, c := range all {
			if c.ID != id && c.Name == in.Name {
				return nil, newAppError(errcode.CodeConflict, "已存在同名配置")
			}
		}
		fields["name"] = in.Name
	}
	if in.Provider != "" {
		if _, ok := allowedProviders[in.Provider]; !ok {
			return nil, newAppError(errcode.CodeBadRequest, "不支持的供应商")
		}
		fields["provider"] = in.Provider
	}
	if in.Model != "" {
		if len(in.Model) > 100 {
			return nil, newAppError(errcode.CodeBadRequest, "模型名称长度需为 1-100")
		}
		fields["model"] = in.Model
	}
	if in.BaseURL != "" {
		if err := validateBaseURL(in.BaseURL); err != nil {
			return nil, newAppError(errcode.CodeBadRequest, err.Error())
		}
		fields["base_url"] = in.BaseURL
	}
	if in.APIKey != "" {
		enc, prefix, suffix, err := s.encryptKey(in.APIKey)
		if err != nil {
			return nil, err
		}
		fields["api_key_enc"] = enc
		fields["api_key_prefix"] = prefix
		fields["api_key_suffix"] = suffix
	}
	if in.IsEnabled != nil {
		fields["is_enabled"] = *in.IsEnabled
	}
	if in.IsDefault != nil {
		fields["is_default"] = *in.IsDefault
	}

	// 默认必须启用（综合 IsEnabled 判断）
	finalEnabled := existing.IsEnabled
	if v, ok := fields["is_enabled"]; ok {
		if b, _ := v.(bool); ok {
			finalEnabled = b
		}
	}
	finalDefault := existing.IsDefault
	if v, ok := fields["is_default"]; ok {
		if b, _ := v.(bool); ok {
			finalDefault = b
		}
	}
	if finalDefault && !finalEnabled {
		return nil, newAppError(errcode.CodeBadRequest, "默认模型必须同时启用")
	}

	if finalDefault {
		_, err := s.repo.WithTx(ctx, func(r repository.LLMConfigRepo) (any, error) {
			if err := r.ClearDefault(ctx, userID); err != nil {
				return nil, err
			}
			if err := r.Update(ctx, userID, id, fields); err != nil {
				return nil, err
			}
			return nil, nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.Update(ctx, userID, id, fields); err != nil {
			return nil, err
		}
	}

	return s.repo.GetByID(ctx, userID, id)
}

// SetDefault 把某条配置设为默认（事务：先清空其它默认，再设当前）。
func (s *LLMConfigService) SetDefault(
	ctx context.Context,
	userID, id uint,
) (*model.LLMConfig, error) {
	cfg, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, newAppError(errcode.CodeNotFound, "大模型配置不存在")
	}
	if !cfg.IsEnabled {
		return nil, newAppError(errcode.CodeBadRequest, "请先启用该配置")
	}

	_, err = s.repo.WithTx(ctx, func(r repository.LLMConfigRepo) (any, error) {
		if err := r.ClearDefault(ctx, userID); err != nil {
			return nil, err
		}
		if err := r.Update(ctx, userID, id, map[string]any{"is_default": true}); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, userID, id)
}

// Delete 删除配置。
func (s *LLMConfigService) Delete(ctx context.Context, userID, id uint) error {
	n, err := s.repo.CountByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if n == 0 {
		return newAppError(errcode.CodeNotFound, "大模型配置不存在")
	}
	return s.repo.Delete(ctx, userID, id)
}

// Get 获取单条。
func (s *LLMConfigService) Get(ctx context.Context, userID, id uint) (*model.LLMConfig, error) {
	cfg, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, newAppError(errcode.CodeNotFound, "大模型配置不存在")
	}
	return cfg, nil
}

// List 列出某 user 的全部配置。
func (s *LLMConfigService) List(ctx context.Context, userID uint) ([]*model.LLMConfig, error) {
	return s.repo.List(ctx, userID)
}

// ListEnabled 列出某 user 的启用配置（给 AI 创建 / 调优使用）。
func (s *LLMConfigService) ListEnabled(ctx context.Context, userID uint) ([]*model.LLMConfig, error) {
	return s.repo.ListEnabled(ctx, userID)
}

// TestConnection 真实 HTTP POST chat/completions，10s timeout。
func (s *LLMConfigService) TestConnection(
	ctx context.Context,
	userID, id uint,
) (*TestResult, error) {
	if s.aes == nil {
		return &TestResult{
			OK:           false,
			ErrorCode:    testCodeNetworkError,
			ErrorMessage: "服务端未配置 LLM 加密密钥（PROMPTHUB_LLM_AES_KEY）",
		}, nil
	}
	cfg, err := s.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	plain, err := s.aes.Decrypt(cfg.APIKeyEnc)
	if err != nil {
		return nil, newAppError(errcode.CodeServiceUnavailable, "密钥解密失败")
	}

	endpoint := strings.TrimRight(cfg.BaseURL, "/") + "/chat/completions"
	body, _ := json.Marshal(map[string]any{
		"model": cfg.Model,
		"messages": []map[string]string{
			{"role": "user", "content": "hi"},
		},
		"max_tokens": 1,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return &TestResult{OK: false, ErrorCode: testCodeNetworkError, ErrorMessage: err.Error()}, nil
	}
	req.Header.Set("Authorization", "Bearer "+plain)
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := s.http.Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		var ne net.Error
		if errors.As(err, &ne) && ne.Timeout() {
			return &TestResult{OK: false, LatencyMs: latency, ErrorCode: testCodeTimeout, ErrorMessage: "连接超时"}, nil
		}
		return &TestResult{OK: false, LatencyMs: latency, ErrorCode: testCodeNetworkError, ErrorMessage: err.Error()}, nil
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body) // 读完 body 让 keep-alive 工作

	switch resp.StatusCode {
	case 200:
		return &TestResult{OK: true, LatencyMs: latency}, nil
	case 401, 403:
		return &TestResult{OK: false, LatencyMs: latency, ErrorCode: testCodeAuthFailed, ErrorMessage: "鉴权失败"}, nil
	case 404:
		return &TestResult{OK: false, LatencyMs: latency, ErrorCode: testCodeModelNotFound, ErrorMessage: "模型不存在或无权访问"}, nil
	case 429:
		return &TestResult{OK: false, LatencyMs: latency, ErrorCode: testCodeQuotaExhausted, ErrorMessage: "配额已用尽"}, nil
	default:
		return &TestResult{OK: false, LatencyMs: latency, ErrorCode: testCodeNetworkError, ErrorMessage: fmt.Sprintf("HTTP %d", resp.StatusCode)}, nil
	}
}

// validateInput 通用校验（create=true 强制 ApiKey 必填）。
func (s *LLMConfigService) validateInput(in LLMConfigInput, create bool) error {
	if create {
		if len(in.Name) == 0 || len(in.Name) > 20 {
			return newAppError(errcode.CodeBadRequest, "名称长度需为 1-20")
		}
		if len(in.Model) == 0 || len(in.Model) > 100 {
			return newAppError(errcode.CodeBadRequest, "模型名称长度需为 1-100")
		}
		if in.APIKey == "" {
			return newAppError(errcode.CodeBadRequest, "API Key 必填")
		}
	} else {
		if in.Name != "" && len(in.Name) > 20 {
			return newAppError(errcode.CodeBadRequest, "名称长度需为 1-20")
		}
		if in.Model != "" && len(in.Model) > 100 {
			return newAppError(errcode.CodeBadRequest, "模型名称长度需为 1-100")
		}
	}
	if in.Provider != "" {
		if _, ok := allowedProviders[in.Provider]; !ok {
			return newAppError(errcode.CodeBadRequest, "不支持的供应商")
		}
	}
	if in.BaseURL != "" {
		if err := validateBaseURL(in.BaseURL); err != nil {
			return newAppError(errcode.CodeBadRequest, err.Error())
		}
	}
	if create || in.APIKey != "" {
		if s.aes == nil {
			return newAppError(errcode.CodeServiceUnavailable, "服务端未配置 LLM 加密密钥")
		}
	}
	return nil
}

// encryptKey 用 AES-GCM 加密 API Key，并拆 prefix（首 6）/suffix（末 4）用于脱敏展示。
func (s *LLMConfigService) encryptKey(plain string) (enc, prefix, suffix string, err error) {
	if s.aes == nil {
		return "", "", "", newAppError(errcode.CodeServiceUnavailable, "服务端未配置 LLM 加密密钥")
	}
	enc, err = s.aes.Encrypt(plain)
	if err != nil {
		return
	}
	// prefix / suffix：明文首 6 / 末 4（用于列表脱敏展示）
	if len(plain) <= 10 {
		prefix = plain
		suffix = ""
	} else {
		prefix = plain[:6]
		suffix = plain[len(plain)-4:]
	}
	return enc, prefix, suffix, nil
}

// validateBaseURL 校验 URL 格式（http/https）。
func validateBaseURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("BaseURL 不是合法 URL")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("BaseURL 必须以 http/https 开头")
	}
	if u.Host == "" {
		return fmt.Errorf("BaseURL 缺少 host")
	}
	return nil
}