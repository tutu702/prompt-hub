// Package middleware 提供 HTTP 中间件。
//
// 现阶段提供：
//   - Recovery：捕获 panic 返回统一 500
//   - AccessLog：访问日志
//   - CORS：跨域处理
package middleware

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/prompthub/prompthub/internal/pkg/response"
)

// Recovery 捕获 panic 并返回 500。
func Recovery(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error("panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.ByteString("stack", debug.Stack()),
				)
				response.FailWithStatus(c, http.StatusInternalServerError, "服务器内部错误")
			}
		}()
		c.Next()
	}
}

// AccessLog 输出访问日志。
func AccessLog(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		if raw != "" {
			path = path + "?" + raw
		}
		log.Info("access",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.String("ip", c.ClientIP()),
			zap.Duration("latency", latency),
			zap.String("ua", c.Request.UserAgent()),
		)
	}
}

// CORS 跨域中间件。生产环境请配置 ORIGIN_ALLOW 白名单。
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
