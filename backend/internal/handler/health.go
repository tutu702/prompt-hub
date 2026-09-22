// Package handler 负责 HTTP 请求的业务处理层（接口层）。
//
// 当前骨架仅提供：
//   - Healthz：健康检查
package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/prompthub/prompthub/internal/config"
	"github.com/prompthub/prompthub/internal/pkg/response"
)

// HealthHandler 健康检查控制器。
type HealthHandler struct {
	cfg *config.Config
	db  *gorm.DB
}

// NewHealthHandler 创建健康检查控制器。
func NewHealthHandler(cfg *config.Config, db *gorm.DB) *HealthHandler {
	return &HealthHandler{cfg: cfg, db: db}
}

// Healthz 返回应用与依赖的整体健康状态。
//
// GET /healthz
func (h *HealthHandler) Healthz(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	dbStatus := "ok"
	if h.db != nil {
		sqlDB, err := h.db.WithContext(ctx).DB()
		if err == nil {
			if err := sqlDB.PingContext(ctx); err != nil {
				dbStatus = "fail: " + err.Error()
			}
		} else {
			dbStatus = "fail: " + err.Error()
		}
	} else {
		dbStatus = "not configured"
	}

	response.OK(c, gin.H{
		"app":       h.cfg.App.Name,
		"version":   h.cfg.App.Version,
		"env":       h.cfg.App.Env,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"deps": gin.H{
			"database": dbStatus,
		},
	})
}

// NotFound 404 处理。
func NotFound(c *gin.Context) {
	response.FailWithStatus(c, http.StatusNotFound, "接口不存在")
}
