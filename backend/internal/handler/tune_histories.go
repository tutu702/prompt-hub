// Package handler - TuneHistoryHandler 处理 /prompts/:id/tune-histories 与 /tune-histories/* 路由（受 JWT 保护）。
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/prompthub/prompthub/internal/handler/dto"
	"github.com/prompthub/prompthub/internal/model"
	"github.com/prompthub/prompthub/internal/pkg/response"
	"github.com/prompthub/prompthub/internal/service"
)

// TuneHistoryHandler 调优历史 handler。
type TuneHistoryHandler struct {
	svc *service.TuneHistoryService
	log *zap.Logger
}

// NewTuneHistoryHandler 构造 TuneHistoryHandler。
func NewTuneHistoryHandler(svc *service.TuneHistoryService, log *zap.Logger) *TuneHistoryHandler {
	return &TuneHistoryHandler{svc: svc, log: log}
}

// Create POST /prompts/:id/tune-histories
func (h *TuneHistoryHandler) Create(c *gin.Context) {
	claims, ok := userClaims(c)
	if !ok {
		return
	}
	promptID, err := parseID(c)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "无效的提示词 ID")
		return
	}
	var req dto.CreateTuneHistoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	history, err := h.svc.Create(c.Request.Context(), claims.UserID, promptID, service.TuneHistoryInput{
		RequestText:     req.RequestText,
		OriginalContent: req.OriginalContent,
		TunedContent:    req.TunedContent,
		ModelConfigID:   req.ModelConfigID,
	})
	if err != nil {
		writeErr(c, h.log, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"code":    200,
		"message": "OK",
		"data":    toTuneHistoryDTO(history),
	})
}

// List GET /prompts/:id/tune-histories
func (h *TuneHistoryHandler) List(c *gin.Context) {
	claims, ok := userClaims(c)
	if !ok {
		return
	}
	promptID, err := parseID(c)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "无效的提示词 ID")
		return
	}
	var req dto.ListTuneHistoriesReq
	_ = c.ShouldBindQuery(&req)
	res, err := h.svc.List(c.Request.Context(), claims.UserID, promptID, req.Page, req.PageSize)
	if err != nil {
		writeErr(c, h.log, err)
		return
	}
	response.Page(c, toTuneHistoryDTOs(res.List), res.Total, res.Page, res.Size)
}

// Get GET /tune-histories/:id
func (h *TuneHistoryHandler) Get(c *gin.Context) {
	claims, ok := userClaims(c)
	if !ok {
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "无效的历史 ID")
		return
	}
	history, err := h.svc.Get(c.Request.Context(), claims.UserID, id)
	if err != nil {
		writeErr(c, h.log, err)
		return
	}
	response.OK(c, toTuneHistoryDTO(history))
}

// Apply POST /tune-histories/:id/apply
func (h *TuneHistoryHandler) Apply(c *gin.Context) {
	claims, ok := userClaims(c)
	if !ok {
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "无效的历史 ID")
		return
	}
	p, err := h.svc.Apply(c.Request.Context(), claims.UserID, id)
	if err != nil {
		writeErr(c, h.log, err)
		return
	}
	// 返回更新后的 prompt，便于前端做本地缓存同步
	response.OK(c, toPromptDTO(p))
}

// --- helpers ---

// toTuneHistoryDTO model → DTO。
func toTuneHistoryDTO(h *model.TuneHistory) dto.TuneHistoryDTO {
	if h == nil {
		return dto.TuneHistoryDTO{}
	}
	return dto.TuneHistoryDTO{
		ID:              h.ID,
		CreatedAt:       h.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       h.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UserID:          h.UserID,
		PromptID:        h.PromptID,
		VersionNo:       h.VersionNo,
		RequestText:     h.RequestText,
		OriginalContent: h.OriginalContent,
		TunedContent:    h.TunedContent,
		ModelConfigID:   h.ModelConfigID,
		IsApplied:       h.IsApplied,
	}
}

// toTuneHistoryDTOs 批量转换。
func toTuneHistoryDTOs(list []*model.TuneHistory) []dto.TuneHistoryDTO {
	out := make([]dto.TuneHistoryDTO, 0, len(list))
	for _, h := range list {
		out = append(out, toTuneHistoryDTO(h))
	}
	return out
}