// Package handler - CategoryHandler 处理 /categories/* 路由（受 JWT 保护）。
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

// CategoryHandler 处理 /categories/* 路由。
type CategoryHandler struct {
	svc *service.CategoryService
	log *zap.Logger
}

// NewCategoryHandler 构造 CategoryHandler。
func NewCategoryHandler(svc *service.CategoryService, log *zap.Logger) *CategoryHandler {
	return &CategoryHandler{svc: svc, log: log}
}

// List GET /categories。
func (h *CategoryHandler) List(c *gin.Context) {
	claims, ok := userClaims(c)
	if !ok {
		return
	}
	list, err := h.svc.List(c.Request.Context(), claims.UserID)
	if err != nil {
		writeErr(c, h.log, err)
		return
	}
	dtos := make([]dto.CategoryDTO, 0, len(list))
	for _, item := range list {
		dtos = append(dtos, toCategoryDTO(item))
	}
	response.OK(c, gin.H{"list": dtos})
}

// Create POST /categories。
func (h *CategoryHandler) Create(c *gin.Context) {
	claims, ok := userClaims(c)
	if !ok {
		return
	}
	var req dto.CreateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	cat, err := h.svc.Create(c.Request.Context(), claims.UserID, service.CategoryInput{
		Name:  req.Name,
		Color: req.Color,
		Sort:  req.Sort,
	})
	if err != nil {
		writeErr(c, h.log, err)
		return
	}
	response.OK(c, toCategoryDTO(cat))
}

// Update PUT /categories/:id。
func (h *CategoryHandler) Update(c *gin.Context) {
	claims, ok := userClaims(c)
	if !ok {
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "无效的分类 ID")
		return
	}
	var req dto.UpdateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	// 把 *string 折叠成可空字符串：nil → "" 视为不改（service 内识别）
	var colorStr string
	if req.Color != nil {
		colorStr = *req.Color
	}
	var nameStr string
	if req.Name != nil {
		nameStr = *req.Name
	}
	cat, err := h.svc.Update(c.Request.Context(), claims.UserID, id, service.CategoryInput{
		Name:  nameStr,
		Color: colorStr,
		Sort:  req.Sort,
	})
	if err != nil {
		writeErr(c, h.log, err)
		return
	}
	response.OK(c, toCategoryDTO(cat))
}

// Delete DELETE /categories/:id。
func (h *CategoryHandler) Delete(c *gin.Context) {
	claims, ok := userClaims(c)
	if !ok {
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "无效的分类 ID")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), claims.UserID, id); err != nil {
		if ae, ok := service.IsAppError(err); ok && ae.Code == errcode.CodeConflict {
			response.Fail(c, http.StatusConflict, ae.Code, ae.Message)
			return
		}
		writeErr(c, h.log, err)
		return
	}
	response.OK(c, gin.H{})
}

// toCategoryDTO model → DTO。
func toCategoryDTO(c *model.Category) dto.CategoryDTO {
	return dto.CategoryDTO{
		ID:        c.ID,
		CreatedAt: c.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt: c.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UserID:    c.UserID,
		Name:      c.Name,
		Color:     c.Color,
		Sort:      c.Sort,
	}
}