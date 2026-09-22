// Package middleware 提供 HTTP 中间件。
//
// jwt.go 实现 JWT 鉴权中间件：从 Authorization Header 解析 Bearer token，
// 成功后将 claims 写入 ctx（key="auth"）；失败统一返回 401。
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/prompthub/prompthub/internal/pkg/jwt"
	"github.com/prompthub/prompthub/internal/pkg/response"
)

// authContextKey 用于在 gin.Context 中存取 *jwt.Claims。
const authContextKey = "auth"

// JWT 构造一个鉴权中间件：解析 Authorization: Bearer <token>。
func JWT(mgr *jwt.Manager, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		if raw == "" {
			response.FailWithStatus(c, http.StatusUnauthorized, "未登录或登录已过期")
			return
		}
		parts := strings.SplitN(raw, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			response.FailWithStatus(c, http.StatusUnauthorized, "未登录或登录已过期")
			return
		}
		claims, err := mgr.Parse(parts[1])
		if err != nil {
			log.Debug("jwt parse failed", zap.Error(err))
			response.FailWithStatus(c, http.StatusUnauthorized, "登录已过期或 token 无效")
			return
		}
		if claims.Subject != "access" {
			response.FailWithStatus(c, http.StatusUnauthorized, "登录已过期或 token 无效")
			return
		}
		c.Set(authContextKey, claims)
		c.Next()
	}
}

// GetAuth 从 gin.Context 中取出 *jwt.Claims。
func GetAuth(c *gin.Context) (*jwt.Claims, bool) {
	v, ok := c.Get(authContextKey)
	if !ok {
		return nil, false
	}
	claims, ok := v.(*jwt.Claims)
	return claims, ok
}