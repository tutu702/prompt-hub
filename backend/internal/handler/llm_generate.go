// Package handler - LLMGenerateHandler 处理 POST /api/v1/llm/generate（SSE 流式响应）。
//
// 协议与 /llm/tune 一致：
//   - 请求：application/json，含 purpose / style? / language? / model_config_id? / mock?
//   - 响应：text/event-stream
//       event: chunk   data: {"delta":"..."}
//       event: done    data: {"result":{"title":"...","content":"...","description":"..."}}
//       event: error   data: {"message":"..."}
//
// 客户端断开通过 ctx.Done 通知；onDelta 返回 false 时 service 立刻停止读上游。
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/prompthub/prompthub/internal/handler/dto"
	"github.com/prompthub/prompthub/internal/pkg/response"
	"github.com/prompthub/prompthub/internal/service"
)

// LLMGenerateHandler 生成 handler。
type LLMGenerateHandler struct {
	svc *service.LLMGenerateService
	log *zap.Logger
}

// NewLLMGenerateHandler 构造 handler。
func NewLLMGenerateHandler(svc *service.LLMGenerateService, log *zap.Logger) *LLMGenerateHandler {
	return &LLMGenerateHandler{svc: svc, log: log}
}

// Generate POST /api/v1/llm/generate
func (h *LLMGenerateHandler) Generate(c *gin.Context) {
	claims, ok := userClaims(c)
	if !ok {
		return
	}
	var req dto.GenerateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// SSE 响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		response.FailWithStatus(c, http.StatusInternalServerError, "SSE 不被支持")
		return
	}

	clientGone := false
	go func() {
		<-c.Request.Context().Done()
		clientGone = true
	}()

	// 立即 flush 一个 comment 行
	_, _ = fmt.Fprintf(c.Writer, ": connected\n\n")
	flusher.Flush()

	sendEvent := func(event string, payload any) bool {
		data, _ := json.Marshal(payload)
		if _, err := fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, data); err != nil {
			return false
		}
		flusher.Flush()
		return true
	}

	result, err := h.svc.Generate(c.Request.Context(), claims.UserID, service.GenerateInput{
		Purpose:       req.Purpose,
		Style:         req.Style,
		Language:      req.Language,
		ModelConfigID: req.ModelConfigID,
		Mock:          req.Mock,
	}, func(delta string) bool {
		if clientGone {
			return false
		}
		return sendEvent("chunk", map[string]string{"delta": delta})
	})

	if clientGone || errors.Is(c.Request.Context().Err(), context.Canceled) {
		return
	}

	if err != nil {
		if ae, ok := service.IsAppError(err); ok {
			sendEvent("error", map[string]any{
				"message": ae.Message,
				"code":    ae.Code,
			})
		} else {
			h.log.Error("llm generate 失败", zap.Error(err))
			sendEvent("error", map[string]string{"message": "生成失败，请稍后再试"})
		}
		return
	}

	sendEvent("done", map[string]any{
		"result": result,
	})
}