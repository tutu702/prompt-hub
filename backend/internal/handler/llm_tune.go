// Package handler - LLMTuneHandler 处理 POST /api/v1/llm/tune（SSE 流式响应）。
//
// 协议：
//   - 请求：application/json，含 prompt_id / request_text / use_context / model_config_id? / mock?
//   - 响应：text/event-stream，三种事件
//       event: chunk   data: {"delta":"..."}
//       event: done    data: {"history":{...TuneHistory}}
//       event: error   data: {"message":"..."}
//
// 客户端断开通过 ctx.Done 通知；onDelta 返回 false 时 service 立刻停止读上游。
//
// 日志：记录进入时的关键参数 + 每个 SSE 事件下发的字节数与计数；
// chunk 不逐条打（避免日志爆炸），仅出汇总：chunk 数 / done / error / 总耗时。
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/prompthub/prompthub/internal/handler/dto"
	"github.com/prompthub/prompthub/internal/pkg/response"
	"github.com/prompthub/prompthub/internal/service"
)

// LLMTuneHandler 调优 handler。
type LLMTuneHandler struct {
	svc *service.LLMTuneService
	log *zap.Logger
}

// NewLLMTuneHandler 构造 handler。
func NewLLMTuneHandler(svc *service.LLMTuneService, log *zap.Logger) *LLMTuneHandler {
	return &LLMTuneHandler{svc: svc, log: log}
}

// Tune POST /api/v1/llm/tune
func (h *LLMTuneHandler) Tune(c *gin.Context) {
	claims, ok := userClaims(c)
	if !ok {
		return
	}
	var req dto.TuneReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 进入时记录请求关键参数（不含 request_text 全文，避免大文本进日志）。
	modelCfgID := uint(0)
	if req.ModelConfigID != nil {
		modelCfgID = *req.ModelConfigID
	}
	reqStart := time.Now()
	baseLogFields := []zap.Field{
		zap.Uint("user_id", claims.UserID),
		zap.Uint("prompt_id", req.PromptID),
		zap.Uint("model_config_id", modelCfgID),
		zap.Int("request_len", len(req.RequestText)),
		zap.Bool("use_context", req.UseContext),
		zap.Bool("mock", req.Mock),
	}
	h.log.Info("llm tune 请求开始", baseLogFields...)

	// SSE 响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no") // 防止 nginx 缓冲
	c.Writer.WriteHeader(http.StatusOK)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		response.FailWithStatus(c, http.StatusInternalServerError, "SSE 不被支持")
		return
	}

	// 客户端断开感知：ctx.Done 即关页面/取消。
	clientGone := false
	go func() {
		<-c.Request.Context().Done()
		clientGone = true
	}()

	// 立即 flush 一个 comment 行，避免某些代理超时
	_, _ = fmt.Fprintf(c.Writer, ": connected\n\n")
	flusher.Flush()

	// SSE 事件下发统计
	var (
		chunkCount   int
		chunkBytes   int
		lastEvent    string
		sendSucceeded = true
	)
	sendEvent := func(event string, payload any) bool {
		data, _ := json.Marshal(payload)
		frame := fmt.Sprintf("event: %s\ndata: %s\n\n", event, data)
		n, err := fmt.Fprint(c.Writer, frame)
		if err != nil {
			sendSucceeded = false
			return false
		}
		flusher.Flush()
		lastEvent = event
		switch event {
		case "chunk":
			chunkCount++
			chunkBytes += n
		}
		h.log.Debug("llm tune sse event",
			append(baseLogFields,
				zap.String("event", event),
				zap.Int("frame_bytes", n),
			)...,
		)
		return true
	}

	history, err := h.svc.Tune(c.Request.Context(), claims.UserID, service.TuneInput{
		PromptID:      req.PromptID,
		RequestText:   req.RequestText,
		UseContext:    req.UseContext,
		ModelConfigID: req.ModelConfigID,
		Mock:          req.Mock,
	}, func(delta string) bool {
		if clientGone {
			return false
		}
		return sendEvent("chunk", map[string]string{"delta": delta})
	})

	if clientGone || errors.Is(c.Request.Context().Err(), context.Canceled) {
		// 客户端断开：不再下发事件
		h.log.Info("llm tune 客户端断开",
			append(baseLogFields,
				zap.Int("chunks", chunkCount),
				zap.Int("chunk_bytes", chunkBytes),
				zap.String("last_event", lastEvent),
				zap.Duration("elapsed", time.Since(reqStart)),
			)...,
		)
		return
	}

	if err != nil {
		if ae, ok := service.IsAppError(err); ok {
			sendEvent("error", map[string]any{
				"message": ae.Message,
				"code":    ae.Code,
			})
			h.log.Warn("llm tune 业务错误",
				append(baseLogFields,
					zap.Int("chunks", chunkCount),
					zap.Int("chunk_bytes", chunkBytes),
					zap.Int("err_code", ae.Code),
					zap.String("err_message", ae.Message),
					zap.Duration("elapsed", time.Since(reqStart)),
				)...,
			)
		} else {
			h.log.Error("llm tune 失败", zap.Error(err))
			sendEvent("error", map[string]string{"message": "调优失败，请稍后再试"})
		}
		return
	}

	sendEvent("done", map[string]any{
		"history": toTuneHistoryDTO(history),
	})
	h.log.Info("llm tune 完成",
		append(baseLogFields,
			zap.Int("chunks", chunkCount),
			zap.Int("chunk_bytes", chunkBytes),
			zap.Uint("history_id", history.ID),
			zap.Int("tuned_len", len(history.TunedContent)),
			zap.Bool("send_ok", sendSucceeded),
			zap.Duration("elapsed", time.Since(reqStart)),
		)...,
	)
}