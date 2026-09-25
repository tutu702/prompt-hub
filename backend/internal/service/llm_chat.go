// Package service - LLMChatClient 封装对大模型 OpenAI 兼容接口的流式调用。
//
// 用途：被 LLMTuneService / LLMGenerateService 共享，避免重复实现 SSE 解析 / 重试 / 错误映射。
//
// 调用流程：
//   1. 上游 POST {base_url}/chat/completions，body {model, messages, stream=true}
//   2. 读响应体按 SSE 帧解析 data: {...}
//   3. 抽 choices[0].delta.content 增量回调 onDelta；返回 false 表示调用方要求中止
//   4. 返回完整拼接文本
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

	"github.com/prompthub/prompthub/internal/model"
	"github.com/prompthub/prompthub/internal/pkg/crypto"
	"github.com/prompthub/prompthub/pkg/errcode"
)

// LLMChatClient 调用上游大模型（OpenAI 兼容 stream）。
type LLMChatClient struct {
	aes  *crypto.AES // nil 时不允许真实调用（503）
	http *http.Client
}

// NewLLMChatClient 构造客户端。http 由调用方注入超时；流式不设硬超时（由 ctx 控制）。
func NewLLMChatClient(aes *crypto.AES) *LLMChatClient {
	return &LLMChatClient{
		aes:  aes,
		http: &http.Client{Timeout: 0},
	}
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", newAppError(errcode.CodeInternal, "构造请求失败: "+err.Error())
	}
	req.Header.Set("Authorization", "Bearer "+plain)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", newAppError(errcode.CodeBadGateway, "上游连接失败: "+err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", newAppError(errcode.CodeBadGateway,
			fmt.Sprintf("上游返回 %d: %s", resp.StatusCode, truncateStr(string(raw), 200)))
	}

	reader := bufio.NewReaderSize(resp.Body, 16<<10)
	var full strings.Builder
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
					delta := parseDelta(payload)
					if delta != "" {
						full.WriteString(delta)
						if !onDelta(delta) {
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
				return full.String(), nil
			}
			return "", newAppError(errcode.CodeBadGateway, "读取上游流失败: "+err.Error())
		}
	}
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