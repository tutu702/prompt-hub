// Package service - LLMTuneService AI 调优业务编排（PRD P07 / F09）。
//
// 职责：
//   1. 校验 prompt 归属与 model_config 归属 + 启用状态；
//   2. 调真实 LLM（OpenAI 兼容 chat/completions，stream=true）或本地 mock 生成；
//   3. 通过 onDelta 回调把片段传给 handler，由 handler 用 SSE 推给前端；
//   4. 流式结束后把完整结果写入 tune_histories（version_no 自增）。
//
// 约束：
//   - LLM 上游调用统一走 LLMChatClient，不在 handler 直接拼 HTTP；
//   - mock 模式下仍走完整流程（写历史），保证本地开发联调体验与生产一致。
package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/prompthub/prompthub/internal/model"
	"github.com/prompthub/prompthub/internal/pkg/crypto"
	"github.com/prompthub/prompthub/internal/repository"
	"github.com/prompthub/prompthub/pkg/errcode"
)

// LLMTuneService 调优服务。
type LLMTuneService struct {
	prompts   repository.PromptRepo
	cfgs      repository.LLMConfigRepo
	histories repository.TuneHistoryRepo
	chat      *LLMChatClient
	log       *zap.Logger
}

// NewLLMTuneService 构造调优服务。aes 为 nil 时 mock 模式仍可用，真实模式返回 503。
func NewLLMTuneService(
	prompts repository.PromptRepo,
	cfgs repository.LLMConfigRepo,
	histories repository.TuneHistoryRepo,
	aes *crypto.AES,
	log *zap.Logger,
) *LLMTuneService {
	return &LLMTuneService{
		prompts:   prompts,
		cfgs:      cfgs,
		histories: histories,
		chat:      NewLLMChatClient(aes),
		log:       log,
	}
}

// TuneInput 调优入参（service 层独立校验）。
type TuneInput struct {
	PromptID      uint
	RequestText   string
	UseContext    bool
	ModelConfigID *uint
	Mock          bool
}

// TuneChunkHandler 流式片段回调；返回 false 表示 handler 请求中止（client 断开），service 应停止读 LLM。
type TuneChunkHandler func(delta string) bool

// Tune 串起完整调优流程：校验 → 调 LLM（流式） → 写历史 → 返回 history。
//
// 调用方（handler）应负责把 onDelta 通过 SSE 推给前端；当 stream 完成后返回的 *model.TuneHistory
// 由 handler 包装到 SSE "done" event 一并下发。
func (l *LLMTuneService) Tune(
	ctx context.Context,
	userID uint,
	in TuneInput,
	onDelta TuneChunkHandler,
) (*model.TuneHistory, error) {
	// 1) prompt 归属校验
	p, err := l.prompts.GetByID(ctx, userID, in.PromptID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, newAppError(errcode.CodeNotFound, "提示词不存在")
	}

	req := strings.TrimSpace(in.RequestText)
	if req == "" {
		return nil, newAppError(errcode.CodeBadRequest, "调优诉求不能为空")
	}

	// 2) model_config 解析：未指定则取 user 默认；未启用则报错
	cfg, err := l.resolveConfig(ctx, userID, in.ModelConfigID)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, newAppError(errcode.CodeBadRequest, "请先在【大模型配置】启用至少一个模型")
	}
	if !cfg.IsEnabled {
		return nil, newAppError(errcode.CodeBadRequest, "所选模型未启用")
	}

	// 3) 调 LLM（mock 或真实）
	var fullText string
	if in.Mock {
		fullText = buildMockTuneOutput(req, p.Content, in.UseContext)
		if !streamMockChunks(fullText, onDelta) {
			return nil, newAppError(errcode.CodeInternal, "客户端断开")
		}
	} else {
		fullText, err = l.chat.Stream(ctx, cfg, l.buildMessages(req, p, in.UseContext), func(delta string) bool {
			return onDelta(delta)
		})
		if err != nil {
			return nil, err
		}
	}

	if strings.TrimSpace(fullText) == "" {
		return nil, newAppError(errcode.CodeBadGateway, "大模型返回内容为空")
	}

	// 4) 写历史
	maxV, err := l.histories.MaxVersionNo(ctx, userID, in.PromptID)
	if err != nil {
		return nil, err
	}
	h := &model.TuneHistory{
		UserID:          userID,
		PromptID:        in.PromptID,
		VersionNo:       maxV + 1,
		RequestText:     req,
		OriginalContent: p.Content,
		TunedContent:    fullText,
		ModelConfigID:   &cfg.ID,
		IsApplied:       false,
	}
	if err := l.histories.Create(ctx, h); err != nil {
		return nil, err
	}
	return h, nil
}

// resolveConfig 解析配置：显式 > 默认 > nil。
func (l *LLMTuneService) resolveConfig(
	ctx context.Context,
	userID uint,
	id *uint,
) (*model.LLMConfig, error) {
	if id != nil {
		c, err := l.cfgs.GetByID(ctx, userID, *id)
		if err != nil {
			return nil, err
		}
		if c == nil {
			return nil, newAppError(errcode.CodeNotFound, "所选模型配置不存在")
		}
		return c, nil
	}
	c, err := l.cfgs.GetDefault(ctx, userID)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// buildMessages 组装 OpenAI 兼容 messages：system 指令 + user 诉求 + 原文 + 上下文。
func (l *LLMTuneService) buildMessages(
	reqText string,
	prompt *model.Prompt,
	useContext bool,
) []map[string]string {
	ctxHint := ""
	if useContext {
		desc := prompt.Description
		if desc != "" {
			ctxHint = fmt.Sprintf("\n\n[分类上下文]\n描述：%s", desc)
		} else {
			ctxHint = "\n\n[分类上下文]\n（无描述）"
		}
	}
	user := fmt.Sprintf(
		"调优诉求：%s%s\n\n原文：\n%s\n\n请基于上述诉求改写原文，只返回改写后的完整正文，不要解释、不要前缀。",
		reqText, ctxHint, prompt.Content,
	)
	return []map[string]string{
		{"role": "system", "content": "你是提示词调优助手。用户会给你原文 + 调优诉求，请只输出改写后的完整正文，保持原语言，不要附加解释。"},
		{"role": "user", "content": user},
	}
}

// buildMockTuneOutput 拼装 mock 输出：行级追加 + 上下文提示。
func buildMockTuneOutput(reqText, original string, useContext bool) string {
	trim := strings.TrimSpace(reqText)
	if trim == "" {
		trim = "优化"
	}
	ctx := ""
	if useContext {
		ctx = "\n\n[上下文已附加：分类与描述]"
	}
	lines := strings.Split(original, "\n")
	tuned := make([]string, 0, len(lines))
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			tuned = append(tuned, l)
		} else {
			tuned = append(tuned, l+"（已针对「"+trim+"」优化）")
		}
	}
	out := strings.Join(tuned, "\n") + ctx + "\n\n—— 调优已完成，请审阅后保存或应用。"
	return strings.TrimSpace(out)
}

// streamMockChunks 把整段切成 1-3 字小段，逐段回调 onDelta；返回 false 即客户端断开。
//
// 用普通 func 类型而非 TuneChunkHandler，便于与 GenerateChunkHandler（结构相同）共享。
func streamMockChunks(text string, onDelta func(string) bool) bool {
	chunks := chunkByRunes(text, 3)
	for _, c := range chunks {
		if !onDelta(c) {
			return false
		}
		time.Sleep(8 * time.Millisecond)
	}
	return true
}

func chunkByRunes(s string, size int) []string {
	runes := []rune(s)
	if len(runes) == 0 {
		return nil
	}
	out := make([]string, 0, len(runes)/size+1)
	for i := 0; i < len(runes); i += size {
		end := i + size
		if end > len(runes) {
			end = len(runes)
		}
		out = append(out, string(runes[i:end]))
	}
	return out
}

// 静默导入 gorm 仅占位，避免某些 linter 误判（service 不直接依赖 gorm 但同包其它文件用了）。
var _ = gorm.Expr("")