// Package handler - prompts 占位路由。
//
// 本期仅提供受保护路由的最小占位实现，用于验证 JWT 鉴权流程；
// 真实的提示词 CRUD / 搜索 / 筛选 / 导出按 PRD §三 P03-P07 在后续迭代实现。
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// PromptsPlaceholder GET /api/v1/prompts 占位响应。
func PromptsPlaceholder(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "OK",
		"data": gin.H{
			"list":      []interface{}{},
			"total":     0,
			"page":      1,
			"page_size": 20,
			"note":      "提示词业务尚未开放，本接口仅用于鉴权验证",
		},
	})
}