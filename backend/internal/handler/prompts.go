// Package handler - PromptHandler 处理 /prompts/* 路由（受 JWT 中间件保护）。
//
// 替换原有 PromptsPlaceholder 占位。
package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/prompthub/prompthub/internal/handler/dto"
	"github.com/prompthub/prompthub/internal/middleware"
	"github.com/prompthub/prompthub/internal/model"
	"github.com/prompthub/prompthub/internal/pkg/jwt"
	"github.com/prompthub/prompthub/internal/pkg/response"
	"github.com/prompthub/prompthub/internal/service"
	"github.com/prompthub/prompthub/pkg/errcode"
)

// PromptHandler 处理 /prompts/* 路由。
type PromptHandler struct {
	svc *service.PromptService
	log *zap.Logger
}

// NewPromptHandler 构造 PromptHandler。
func NewPromptHandler(svc *service.PromptService, log *zap.Logger) *PromptHandler {
	return &PromptHandler{svc: svc, log: log}
}

// userClaims 提取当前请求的用户 ID；若 JWT 中间件未挂载则返回 401。
func userClaims(c *gin.Context) (*jwt.Claims, bool) {
	claims, ok := middleware.GetAuth(c)
	if !ok || claims == nil {
		response.FailWithStatus(c, http.StatusUnauthorized, "未登录或登录已过期")
		return nil, false
	}
	return claims, true
}

// List GET /prompts。
func (h *PromptHandler) List(c *gin.Context) {
	claims, ok := userClaims(c)
	if !ok {
		return
	}
	var req dto.ListPromptsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	res, err := h.svc.List(c.Request.Context(), claims.UserID, service.PromptListQuery{
		Keyword:    req.Keyword,
		QIn:        req.QIn,
		CategoryID: req.CategoryID,
		IsFavorite: req.IsFavorite,
		Sort:       req.Sort,
		Page:       req.Page,
		PageSize:   req.PageSize,
	})
	if err != nil {
		writeErr(c, h.log, err)
		return
	}
	response.Page(c, toPromptDTOs(res.List), res.Total, res.Page, res.Size)
}

// Get GET /prompts/:id。
func (h *PromptHandler) Get(c *gin.Context) {
	claims, ok := userClaims(c)
	if !ok {
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "无效的提示词 ID")
		return
	}
	p, err := h.svc.Get(c.Request.Context(), claims.UserID, id)
	if err != nil {
		writeErr(c, h.log, err)
		return
	}
	response.OK(c, toPromptDTO(p))
}

// Create POST /prompts。
func (h *PromptHandler) Create(c *gin.Context) {
	claims, ok := userClaims(c)
	if !ok {
		return
	}
	var req dto.CreatePromptReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	p, err := h.svc.Create(c.Request.Context(), claims.UserID, service.PromptInput{
		Title:       req.Title,
		Description: req.Description,
		Content:     req.Content,
		CategoryID:  req.CategoryID,
		Tags:        req.Tags,
	})
	if err != nil {
		writeErr(c, h.log, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"code":    200,
		"message": "OK",
		"data":    toPromptDTO(p),
	})
}

// Update PUT /prompts/:id。
func (h *PromptHandler) Update(c *gin.Context) {
	claims, ok := userClaims(c)
	if !ok {
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "无效的提示词 ID")
		return
	}
	var req dto.UpdatePromptReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	p, err := h.svc.Update(c.Request.Context(), claims.UserID, id, service.PromptUpdateInput{
		Title:       req.Title,
		Description: req.Description,
		Content:     req.Content,
		CategoryID:  req.CategoryID,
		Tags:        req.Tags,
		IsFavorite:  req.IsFavorite,
	})
	if err != nil {
		writeErr(c, h.log, err)
		return
	}
	response.OK(c, toPromptDTO(p))
}

// Delete DELETE /prompts/:id。
func (h *PromptHandler) Delete(c *gin.Context) {
	claims, ok := userClaims(c)
	if !ok {
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "无效的提示词 ID")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), claims.UserID, id); err != nil {
		writeErr(c, h.log, err)
		return
	}
	response.OK(c, gin.H{})
}

// BatchDelete POST /prompts/batch-delete。
func (h *PromptHandler) BatchDelete(c *gin.Context) {
	claims, ok := userClaims(c)
	if !ok {
		return
	}
	var req dto.BatchDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	n, err := h.svc.BatchDelete(c.Request.Context(), claims.UserID, req.IDs)
	if err != nil {
		writeErr(c, h.log, err)
		return
	}
	response.OK(c, gin.H{"deleted": n})
}

// ToggleFavorite POST /prompts/:id/favorite。
func (h *PromptHandler) ToggleFavorite(c *gin.Context) {
	claims, ok := userClaims(c)
	if !ok {
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "无效的提示词 ID")
		return
	}
	var req dto.ToggleFavoriteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	p, err := h.svc.ToggleFavorite(c.Request.Context(), claims.UserID, id, req.IsFavorite)
	if err != nil {
		writeErr(c, h.log, err)
		return
	}
	response.OK(c, gin.H{"is_favorite": p.IsFavorite})
}

// --- helpers ---

// parseID 解析路径参数 :id；非数字时返回错误。
func parseID(c *gin.Context) (uint, error) {
	raw := c.Param("id")
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New("invalid id")
	}
	return uint(id), nil
}

// writeErr 统一把 service 错误转为响应。
func writeErr(c *gin.Context, log *zap.Logger, err error) {
	if ae, ok := service.IsAppError(err); ok {
		httpStatus := errcode.HTTPStatus(ae.Code)
		response.Fail(c, httpStatus, ae.Code, ae.Message)
		return
	}
	if errors.Is(err, jwt.ErrTokenInvalid) || errors.Is(err, jwt.ErrTokenExpired) {
		response.Fail(c, http.StatusUnauthorized, 401, "登录已过期或 token 无效")
		return
	}
	log.Error("service error", zap.Error(err))
	response.FailWithStatus(c, http.StatusInternalServerError, "服务器内部错误")
}

// toPromptDTO 把 model.Prompt 转 DTO。
func toPromptDTO(p *model.Prompt) dto.PromptDTO {
	if p == nil {
		return dto.PromptDTO{}
	}
	return dto.PromptDTO{
		ID:          p.ID,
		CreatedAt:   p.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   p.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UserID:      p.UserID,
		CategoryID:  p.CategoryID,
		Title:       p.Title,
		Description: p.Description,
		Content:     p.Content,
		IsFavorite:  p.IsFavorite,
		Tags:        service.UnmarshalTags(p.Tags),
		VersionNo:   p.VersionNo,
	}
}

// toPromptDTOs 批量转换。
func toPromptDTOs(list []*model.Prompt) []dto.PromptDTO {
	out := make([]dto.PromptDTO, 0, len(list))
	for _, p := range list {
		out = append(out, toPromptDTO(p))
	}
	return out
}