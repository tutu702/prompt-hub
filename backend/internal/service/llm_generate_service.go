// Package service - LLMGenerateService AI 生成业务编排（PRD F08 / P04 模式 B）。
//
// 职责：
//   1. 解析 model_config（显式 > 默认；未启用报错）；
//   2. 调真实 LLM（OpenAI 兼容 chat/completions，stream=true）或本地 mock 生成；
//   3. 通过 onDelta 回调把片段传给 handler，由 handler 用 SSE 推给前端；
//   4. 流式结束后解析 fullText 为结构化结果 {title, content, description}，通过 handler 下发。
//
// 系统 prompt 约定 LLM 输出格式：
//   - 第 1 行：标题（不超过 20 字）
//   - 第 2 行：空行
//   - 后续：完整正文
//
// 解析策略：trim 后按首个 \n\n 拆分；首段做标题（截前 20 字），余下做 content。
// description 本期由 LLM 不返回，统一空串（V1.1 扩展）。
package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/prompthub/prompthub/internal/model"
	"github.com/prompthub/prompthub/internal/pkg/crypto"
	"github.com/prompthub/prompthub/internal/repository"
	"github.com/prompthub/prompthub/pkg/errcode"
)

// GenerateResult 生成的最终结构化结果。
type GenerateResult struct {
	Title       string `json:"title"`
	Content     string `json:"content"`
	Description string `json:"description"`
}

// LLMGenerateService AI 生成服务。
type LLMGenerateService struct {
	cfgs repository.LLMConfigRepo
	chat *LLMChatClient
	log  *zap.Logger
}

// NewLLLMGenerateService 构造生成服务（命名规避与外部 LLMGenerateService 冲突）。
func NewLLMGenerateService(
	cfgs repository.LLMConfigRepo,
	aes *crypto.AES,
	log *zap.Logger,
) *LLMGenerateService {
	return &LLMGenerateService{
		cfgs: cfgs,
		chat: NewLLMChatClient(aes, log),
		log:  log,
	}
}

// GenerateInput 生成入参。
type GenerateInput struct {
	Purpose       string
	Style         string
	Language      string
	ModelConfigID *uint
	Mock          bool
}

// GenerateChunkHandler 流式片段回调；返回 false 表示客户端断开。
type GenerateChunkHandler func(delta string) bool

// Generate 串起生成流程：校验 → 调 LLM（流式） → 解析 → 返回结构化结果。
//
// 解析逻辑见 parseGenerateOutput。
func (s *LLMGenerateService) Generate(
	ctx context.Context,
	userID uint,
	in GenerateInput,
	onDelta GenerateChunkHandler,
) (*GenerateResult, error) {
	purpose := strings.TrimSpace(in.Purpose)
	if purpose == "" {
		return nil, newAppError(errcode.CodeBadRequest, "用途描述不能为空")
	}

	cfg, err := s.resolveConfig(ctx, userID, in.ModelConfigID)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, newAppError(errcode.CodeBadRequest, "请先在【大模型配置】启用至少一个模型")
	}
	if !cfg.IsEnabled {
		return nil, newAppError(errcode.CodeBadRequest, "所选模型未启用")
	}

	var fullText string
	if in.Mock {
		fullText = buildMockGenerateOutput(purpose, in.Style, in.Language)
		if !streamMockChunks(fullText, onDelta) {
			return nil, newAppError(errcode.CodeInternal, "客户端断开")
		}
	} else {
		fullText, err = s.chat.Stream(ctx, cfg, s.buildMessages(purpose, in.Style, in.Language), func(delta string) bool {
			return onDelta(delta)
		})
		if err != nil {
			return nil, err
		}
	}

	if strings.TrimSpace(fullText) == "" {
		return nil, newAppError(errcode.CodeBadGateway, "大模型返回内容为空")
	}
	return parseGenerateOutput(fullText), nil
}

// resolveConfig 解析配置：显式 > 默认 > nil。
func (s *LLMGenerateService) resolveConfig(
	ctx context.Context,
	userID uint,
	id *uint,
) (*model.LLMConfig, error) {
	if id != nil {
		c, err := s.cfgs.GetByID(ctx, userID, *id)
		if err != nil {
			return nil, err
		}
		if c == nil {
			return nil, newAppError(errcode.CodeNotFound, "所选模型配置不存在")
		}
		return c, nil
	}
	c, err := s.cfgs.GetDefault(ctx, userID)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// buildMessages 组装 messages。
//
// system prompt 明确要求输出格式：第一行标题、第二行空行、后续正文。
// style / language 在 user 段以可读句子注入，模型容易遵循。
func (s *LLMGenerateService) buildMessages(
	purpose, style, language string,
) []map[string]string {
	parts := []string{fmt.Sprintf("用途：%s", purpose)}
	if style != "" {
		parts = append(parts, fmt.Sprintf("风格：%s", styleLabel(style)))
	}
	if language != "" && language != "auto" {
		parts = append(parts, fmt.Sprintf("输出语言：%s", languageLabel(language)))
	}

	sysContent := "你是一名专业提示词编写助手。根据用户的用途、风格与语言偏好，输出可直接复用的提示词正文。" +
		"严格按以下格式：\n" +
		"第 1 行：标题（不超过 20 个字符，简明扼要）。\n" +
		"第 2 行：空行。\n" +
		"第 3 行起：完整可用的提示词正文。\n" +
		"不要附加任何解释、前缀或后缀；不要使用代码块包裹。"

	return []map[string]string{
		{"role": "system", "content": sysContent},
		{"role": "user", "content": strings.Join(parts, "\n")},
	}
}

// parseGenerateOutput 按"第一行标题 + 空行 + 正文"解析 LLM 输出。
//
// 兜底：若 LLM 没遵循格式（无空行分隔），把整个文本当作 content，标题留空。
func parseGenerateOutput(full string) *GenerateResult {
	full = strings.TrimSpace(full)
	if full == "" {
		return &GenerateResult{}
	}
	// 优先按 "\n\n" 拆；找不到则按第一个 "\n" 拆（标题兜底）
	idx := strings.Index(full, "\n\n")
	if idx == -1 {
		idx = strings.Index(full, "\n")
		if idx == -1 {
			// 单行无换行：把全文作为 content，标题空
			return &GenerateResult{Content: full}
		}
	}
	title := strings.TrimSpace(full[:idx])
	content := strings.TrimSpace(full[idx+1:])
	if idx == 0 {
		// 空标题
		title = ""
	}
	// 标题过长截断到 20 字（与前端标题字段长度一致）
	runes := []rune(title)
	if len(runes) > 20 {
		title = string(runes[:20])
	}
	return &GenerateResult{
		Title:       title,
		Content:     content,
		Description: "",
	}
}

// buildMockGenerateOutput 生成 mock 输出（按格式规范）。
func buildMockGenerateOutput(purpose, style, language string) string {
	title := mockTitle(purpose)
	lines := []string{
		"# 角色",
		"你是一名经验丰富的助手。",
		"",
		"# 任务",
		fmt.Sprintf("围绕【%s】提供专业、结构化的输出。", purpose),
		"",
		"# 要求",
		"- 输出语言清晰、逻辑严谨",
		"- 关键步骤用编号或要点列出",
		"- 必要时给出可执行的示例",
	}
	if style != "" {
		lines = append(lines, "", "## 风格偏好", "请以"+styleLabel(style)+"的方式回答。")
	}
	if language != "" && language != "auto" {
		lines = append(lines, "", "## 语言", "请使用"+languageLabel(language)+"输出。")
	}
	lines = append(lines, "", "—— 以上为根据用途自动生成的提示词草稿，可继续调整。")
	return title + "\n\n" + strings.Join(lines, "\n")
}

// mockTitle 从用途描述里粗略抽一个不超过 20 字的标题。
func mockTitle(purpose string) string {
	runes := []rune(purpose)
	if len(runes) > 20 {
		return string(runes[:20])
	}
	return purpose
}

func styleLabel(s string) string {
	switch s {
	case "professional":
		return "专业严谨"
	case "friendly":
		return "友好亲切"
	case "concise":
		return "简洁直接"
	case "detailed":
		return "详尽充分"
	default:
		return s
	}
}

func languageLabel(l string) string {
	switch l {
	case "zh":
		return "中文"
	case "en":
		return "英文"
	default:
		return l
	}
}

// 静默引用 time 以避免 linter 在某些重构后误判（mock 流式 sleep 使用）。
var _ = time.Millisecond