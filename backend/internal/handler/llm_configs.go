// Package handler - LLMConfigHandler 处理 /llm-configs/* 路由（受 JWT 保护）。
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/prompthub/prompthub/internal/handler/dto"
	"github.com/prompthub/prompthub/internal/model"
	"github.com/prompthub/prompthub/internal/pkg/response"
	"github.com/prompthub/prompthub/internal/service"
	"github.com/prompthub/prompthub/pkg/errcode"
)

// LLMConfigHandler 大模型配置 handler。
type LLMConfigHandler struct {
	svc *service.LLMConfigService
	log *zap.Logger
}

// NewLLMConfigHandler 构造 LLMConfigHandler。
func NewLLMConfigHandler(svc *service.LLMConfigService, log *zap.Logger) *LLMConfigHandler {
	return &LLMConfigHandler{svc: svc, log: log}
}

// providerLabelMap 供应商代码 → 友好显示名（前端只展示 ProviderLabel）。
var providerLabelMap = map[string]string{
	"openai":    "OpenAI",
	"anthropic": "Anthropic",
	"deepseek":  "DeepSeek",
	"qwen":      "通义千问",
	"zhipu":     "智谱",
	"moonshot":  "Moonshot",
	"custom":    "自定义",
}

// maskAPIKey 用 prefix + "***" + suffix 拼装脱敏串；prefix 为空时 fallback "***"。
func maskAPIKey(prefix, suffix string) string {
	if prefix == "" {
		return "***"
	}
	return prefix + "***" + suffix
}

// toLLMConfigDTO model → DTO（永远脱敏）。
func toLLMConfigDTO(cfg *model.LLMConfig) dto.LLMConfigDTO {
	return dto.LLMConfigDTO{
		ID:            cfg.ID,
		CreatedAt:     cfg.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     cfg.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UserID:        cfg.UserID,
		Name:          cfg.Name,
		Provider:      cfg.Provider,
		ProviderLabel: providerLabelMap[cfg.Provider],
		Model:         cfg.Model,
		BaseURL:       cfg.BaseURL,
		APIKeyMasked:  maskAPIKey(cfg.APIKeyPrefix, cfg.APIKeySuffix),
		IsEnabled:     cfg.IsEnabled,
		IsDefault:     cfg.IsDefault,
	}
}

// toLLMConfigDTOs 批量转换。
func toLLMConfigDTOs(list []*model.LLMConfig) []dto.LLMConfigDTO {
	out := make([]dto.LLMConfigDTO, 0, len(list))
	for _, c := range list {
		out = append(out, toLLMConfigDTO(c))
	}
	return out
}

// Create POST /llm-configs
func (h *LLMConfigHandler) Create(c *gin.Context) {
	claims, ok := userClaims(c)
	if !ok {
		return
	}
	var req dto.CreateLLMConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	cfg, err := h.svc.Create(c.Request.Context(), claims.UserID, service.LLMConfigInput{
		Name:      req.Name,
		Provider:  req.Provider,
		Model:     req.Model,
		BaseURL:   req.BaseURL,
		APIKey:    req.APIKey,
		IsEnabled: req.IsEnabled,
		IsDefault: req.IsDefault,
	})
	if err != nil {
		writeErr(c, h.log, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"code":    200,
		"message": "OK",
		"data":    toLLMConfigDTO(cfg),
	})
}

// List GET /llm-configs
func (h *LLMConfigHandler) List(c *gin.Context) {
	claims, ok := userClaims(c)
	if !ok {
		return
	}
	list, err := h.svc.List(c.Request.Context(), claims.UserID)
	if err != nil {
		writeErr(c, h.log, err)
		return
	}
	response.OK(c, gin.H{"list": toLLMConfigDTOs(list)})
}

// Get GET /llm-configs/:id
func (h *LLMConfigHandler) Get(c *gin.Context) {
	claims, ok := userClaims(c)
	if !ok {
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "无效的配置 ID")
		return
	}
	cfg, err := h.svc.Get(c.Request.Context(), claims.UserID, id)
	if err != nil {
		writeErr(c, h.log, err)
		return
	}
	response.OK(c, toLLMConfigDTO(cfg))
}

// Update PUT /llm-configs/:id
func (h *LLMConfigHandler) Update(c *gin.Context) {
	claims, ok := userClaims(c)
	if !ok {
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "无效的配置 ID")
		return
	}
	var req dto.UpdateLLMConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	cfg, err := h.svc.Update(c.Request.Context(), claims.UserID, id, service.LLMConfigInput{
		Name:      stringPtrToStr(req.Name),
		Provider:  stringPtrToStr(req.Provider),
		Model:     stringPtrToStr(req.Model),
		BaseURL:   stringPtrToStr(req.BaseURL),
		APIKey:    stringPtrToStr(req.APIKey),
		IsEnabled: req.IsEnabled,
		IsDefault: req.IsDefault,
	})
	if err != nil {
		writeErr(c, h.log, err)
		return
	}
	response.OK(c, toLLMConfigDTO(cfg))
}

// Delete DELETE /llm-configs/:id
func (h *LLMConfigHandler) Delete(c *gin.Context) {
	claims, ok := userClaims(c)
	if !ok {
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "无效的配置 ID")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), claims.UserID, id); err != nil {
		writeErr(c, h.log, err)
		return
	}
	response.OK(c, gin.H{})
}

// SetDefault POST /llm-configs/:id/default
func (h *LLMConfigHandler) SetDefault(c *gin.Context) {
	claims, ok := userClaims(c)
	if !ok {
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "无效的配置 ID")
		return
	}
	cfg, err := h.svc.SetDefault(c.Request.Context(), claims.UserID, id)
	if err != nil {
		writeErr(c, h.log, err)
		return
	}
	response.OK(c, toLLMConfigDTO(cfg))
}

// TestConnection POST /llm-configs/:id/test-connection
func (h *LLMConfigHandler) TestConnection(c *gin.Context) {
	claims, ok := userClaims(c)
	if !ok {
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "无效的配置 ID")
		return
	}
	res, err := h.svc.TestConnection(c.Request.Context(), claims.UserID, id)
	if err != nil {
		writeErr(c, h.log, err)
		return
	}
	resp := dto.TestConnectionResp{
		OK:           res.OK,
		LatencyMs:    res.LatencyMs,
		ErrorCode:    res.ErrorCode,
		ErrorMessage: res.ErrorMessage,
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// stringPtrToStr 把 *string 安全转为 string（nil → ""）。
func stringPtrToStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// 错误兜底：service 返 503 时（如 AES 未配置）writeErr 会按 CodeInternal 返；
// 这里增强一下：service 把"未启用加密"映射到 errcode.CodeConflict 不合适，
// 改用统一的 503 兜底（response 直接 JSON 503）。
func init() {
	// 占位：保证 service 包内 errcode 别名被引用
	_ = errcode.CodeInternal
}