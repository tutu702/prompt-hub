// Package service - LLMChatClient 封装对大模型 OpenAI 兼容接口的流式调用。
//
// 用途：被 LLMTuneService / LLMGenerateService 共享，避免重复实现 SSE 解析 / 重试 / 错误映射。
//
// 调用流程：
//   1. 上游 POST {base_url}/chat/completions，body {model, messages, stream=true}
//   2. 边读边解析 SSE：data: {...} → 回调 onDelta；返回 false 表示调用方要求中止
//   3. 顺便把读到的字节拼到 leftoverBuf：若整段响应里没有任何 data: 帧（上游 200 + JSON 错误体），
//      把 leftoverBuf 内容作为错误信息透出给前端，避免"大模型返回内容为空"这种对用户毫无价值的提示。
//   4. 返回完整拼接文本
//
// 日志约定：
//   - LLM 上游请求 / 响应：endpoint / provider / model / status / 首个 chunk 延迟 / chunk 数 / 最终长度 / 总耗时；
//   - API Key 绝不打日志；messages 仅记条目数和字符长度（不记原文）；
//   - 流式片段逐条不打日志（避免日志爆炸），仅在流结束时出汇总（chunk 数 / 最终长度）。
package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/prompthub/prompthub/internal/model"
	"github.com/prompthub/prompthub/internal/pkg/crypto"
	"github.com/prompthub/prompthub/pkg/errcode"
)

// LLMChatClient 调用上游大模型（OpenAI 兼容 stream）。
type LLMChatClient struct {
	aes  *crypto.AES // nil 时不允许真实调用（503）
	http *http.Client
	log  *zap.Logger // nil 时静默 noop，不阻断正常调用
}

// NewLLMChatClient 构造客户端。http 由调用方注入超时；流式不设硬超时（由 ctx 控制）。
func NewLLMChatClient(aes *crypto.AES, log *zap.Logger) *LLMChatClient {
	return &LLMChatClient{
		aes:  aes,
		http: &http.Client{Timeout: 0},
		log:  log,
	}
}

// messageLengths 统计 messages 每段（role + content）的字符长度，用于日志（不打原文）。
func messageLengths(messages []map[string]string) []int {
	out := make([]int, 0, len(messages))
	for _, m := range messages {
		out = append(out, len(m["role"])+len(m["content"]))
	}
	return out
}

// Stream 调上游 LLM 流式 chat completion。
//
//   - cfg:       已校验的模型配置（含明文 API Key 不在 cfg 内，由本函数内部解密）
//   - messages:  OpenAI 格式 messages
//   - onDelta:   增量片段回调；返回 false 时立刻停止读上游（用于客户端断开场景）
//
// 返回拼接的完整文本。
func (c *LLMChatClient) Stream(
	ctx context.Context,
	cfg *model.LLMConfig,
	messages []map[string]string,
	onDelta func(string) bool,
) (string, error) {
	if c.aes == nil {
		return "", newAppError(errcode.CodeServiceUnavailable,
			"服务端未配置 LLM 加密密钥（PROMPTHUB_LLM_AES_KEY）")
	}
	plain, err := c.aes.Decrypt(cfg.APIKeyEnc)
	if err != nil {
		return "", newAppError(errcode.CodeServiceUnavailable, "密钥解密失败")
	}

	body, _ := json.Marshal(map[string]any{
		"model":       cfg.Model,
		"messages":    messages,
		"stream":      true,
		"temperature": 0.7,
	})
	endpoint := strings.TrimRight(cfg.BaseURL, "/") + "/chat/completions"

	// 日志字段：endpoint / provider / model / messages 数与长度 / body 字节数。
	// 敏感信息（API Key、messages 原文）不入日志。
	baseLogFields := []zap.Field{
		zap.String("endpoint", endpoint),
		zap.String("provider", cfg.Provider),
		zap.String("model", cfg.Model),
		zap.Int("messages", len(messages)),
		zap.Ints("message_lens", messageLengths(messages)),
		zap.Int("body_bytes", len(body)),
	}
	c.log.Info("llm upstream request", baseLogFields...)

	reqStart := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		c.log.Error("llm upstream 构造请求失败",
			append(baseLogFields, zap.Error(err))...,
		)
		return "", newAppError(errcode.CodeInternal, "构造请求失败: "+err.Error())
	}
	req.Header.Set("Authorization", "Bearer "+plain)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.http.Do(req)
	if err != nil {
		c.log.Error("llm upstream 连接失败",
			append(baseLogFields,
				zap.Duration("elapsed", time.Since(reqStart)),
				zap.Error(err),
			)...,
		)
		return "", newAppError(errcode.CodeBadGateway, "上游连接失败: "+err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		c.log.Error("llm upstream 非 200 响应",
			append(baseLogFields,
				zap.Int("status", resp.StatusCode),
				zap.Duration("elapsed", time.Since(reqStart)),
				zap.String("resp_body", truncateStr(string(raw), 500)),
			)...,
		)
		return "", newAppError(errcode.CodeBadGateway,
			fmt.Sprintf("上游返回 %d: %s", resp.StatusCode, truncateStr(string(raw), 200)))
	}

	// 边读边解析：SSE 是行分隔的协议，每读一行就尝试消费一个 data: 帧。
	// 不要先 io.ReadAll 整个 body —— 否则上游流式响应会一直等到结束才向客户端推送 chunk，
	// 用户看到的现象就是「: connected 之后长时间空白」。
	//
	// 用 TeeReader 把读到的字节同时镜像到 leftoverBuf（限 4MB，防恶意 / 异常上游撑爆内存）：
	//   - 正常流式：leftoverBuf 含完整响应体，仅用于日志 / 排查
	//   - 上游返回 200 + 非 SSE（鉴权失败 / 限流时一些 OpenAI 兼容服务返回 200 + JSON 错误体）：
	//     整段读完仍没有任何 data: 帧，把 leftoverBuf 内容作为错误信息透出给前端
	const leftoverCap = 4 << 20
	leftoverBuf := &bytes.Buffer{}
	limitedSrc := io.LimitReader(resp.Body, leftoverCap)
	teeReader := io.TeeReader(limitedSrc, leftoverBuf)
	reader := bufio.NewReaderSize(teeReader, 16<<10)

	var full strings.Builder
	hasData := false
	chunkCount := 0
	var firstChunkAt time.Time // 首个非空 delta 到达时间（用于计算"首个 chunk 延迟"）
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			line = strings.TrimRight(line, "\r\n")
			if strings.HasPrefix(line, "data:") {
				payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
				if payload == "" {
					// 空行：SSE 帧结束
				} else if payload == "[DONE]" {
					break
				} else {
					hasData = true
					delta := parseDelta(payload)
					if delta != "" {
						if firstChunkAt.IsZero() {
							firstChunkAt = time.Now()
						}
						chunkCount++
						full.WriteString(delta)
						if !onDelta(delta) {
							c.log.Info("llm upstream 客户端断开，已停止读上游",
								append(baseLogFields,
									zap.Int("chunks", chunkCount),
									zap.Int("bytes", full.Len()),
									zap.Duration("elapsed", time.Since(reqStart)),
								)...,
							)
							return full.String(), nil
						}
					}
				}
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			if ctx.Err() != nil {
				c.log.Warn("llm upstream ctx 取消",
					append(baseLogFields,
						zap.Int("chunks", chunkCount),
						zap.Int("bytes", full.Len()),
						zap.Duration("elapsed", time.Since(reqStart)),
					)...,
				)
				return full.String(), nil
			}
			c.log.Error("llm upstream 读取流失败",
				append(baseLogFields,
					zap.Int("chunks", chunkCount),
					zap.Int("bytes", full.Len()),
					zap.Duration("elapsed", time.Since(reqStart)),
					zap.Error(err),
				)...,
			)
			return "", newAppError(errcode.CodeBadGateway, "读取上游流失败: "+err.Error())
		}
	}
	// 200 OK 但没有任何 data: 帧 —— 上游返回的并不是 SSE。把镜像的 body 透出，
	// 避免"大模型返回内容为空"这种对用户毫无价值的提示。
	if !hasData {
		c.log.Warn("llm upstream 未返回任何 data 帧",
			append(baseLogFields,
				zap.Int("status", resp.StatusCode),
				zap.Duration("elapsed", time.Since(reqStart)),
				zap.Int("resp_body_bytes", leftoverBuf.Len()),
				zap.String("resp_body_preview", truncateStr(leftoverBuf.String(), 300)),
			)...,
		)
		if msg := strings.TrimSpace(leftoverBuf.String()); msg != "" {
			return "", newAppError(errcode.CodeBadGateway,
				fmt.Sprintf("上游未返回流式数据: %s", truncateStr(msg, 500)))
		}
		return "", newAppError(errcode.CodeBadGateway, "上游未返回流式数据")
	}
	if strings.TrimSpace(full.String()) == "" {
		c.log.Warn("llm upstream 返回内容为空",
			append(baseLogFields,
				zap.Int("chunks", chunkCount),
				zap.Duration("elapsed", time.Since(reqStart)),
			)...,
		)
		return "", newAppError(errcode.CodeBadGateway, "大模型返回内容为空")
	}
	// 成功汇总：上游首字节延迟 / chunk 数 / 文本长度 / 总耗时
	successFields := append([]zap.Field(nil), baseLogFields...)
	successFields = append(successFields,
		zap.Int("status", resp.StatusCode),
		zap.Int("chunks", chunkCount),
		zap.Int("bytes", full.Len()),
	)
	if !firstChunkAt.IsZero() {
		successFields = append(successFields,
			zap.Duration("ttfb", firstChunkAt.Sub(reqStart)),
		)
	}
	successFields = append(successFields, zap.Duration("elapsed", time.Since(reqStart)))
	c.log.Info("llm upstream response done", successFields...)
	return full.String(), nil
}

// parseDelta 解析 OpenAI 流式 chunk，抽取 choices[0].delta.content。
func parseDelta(payload string) string {
	var chunk struct {
		Choices []struct {
			Delta struct {
				Content string `json:"content"`
			} `json:"delta"`
		} `json:"choices"`
	}
	if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
		return ""
	}
	if len(chunk.Choices) == 0 {
		return ""
	}
	return chunk.Choices[0].Delta.Content
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}