// Package handler 负责 HTTP 请求的业务处理层（接口层）。
//
// handler 层只做协议转换：
//   1. 绑定入参（c.ShouldBindJSON / ShouldBindQuery）；
//   2. 调用 service；
//   3. 透传业务错误或包装为统一响应。
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/prompthub/prompthub/internal/handler/dto"
	"github.com/prompthub/prompthub/internal/middleware"
	"github.com/prompthub/prompthub/internal/model"
	"github.com/prompthub/prompthub/internal/pkg/jwt"
	"github.com/prompthub/prompthub/internal/pkg/response"
	"github.com/prompthub/prompthub/internal/service"
)

// AuthHandler 处理 /auth/* 路由。
type AuthHandler struct {
	svc *service.AuthService
	log *zap.Logger
}

// NewAuthHandler 构造 AuthHandler。
func NewAuthHandler(svc *service.AuthService, log *zap.Logger) *AuthHandler {
	return &AuthHandler{svc: svc, log: log}
}

// Register 处理 POST /api/v1/auth/register。
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	res, err := h.svc.Register(c.Request.Context(), service.RegisterInput{
		Email:           req.Email,
		Password:        req.Password,
		ConfirmPassword: req.ConfirmPassword,
		Nickname:        req.Nickname,
		Agreement:       req.Agreement,
	})
	if err != nil {
		h.writeAuthErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "OK",
		"data":    toAuthResultDTO(res),
	})
}

// Login 处理 POST /api/v1/auth/login。
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	res, err := h.svc.Login(c.Request.Context(), service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
		Remember: req.Remember,
	})
	if err != nil {
		h.writeAuthErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "OK",
		"data":    toAuthResultDTO(res),
	})
}

// Refresh 处理 POST /api/v1/auth/refresh。
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	res, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.writeAuthErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "OK",
		"data": gin.H{
			"user":          toAuthUserDTO(res.User),
			"access_token":  res.AccessToken,
			"refresh_token": res.RefreshToken,
			"expires_in":    res.ExpiresIn,
		},
	})
}

// Logout 处理 POST /api/v1/auth/logout（需 JWT 中间件前置）。
func (h *AuthHandler) Logout(c *gin.Context) {
	claims, ok := middleware.GetAuth(c)
	if !ok {
		response.FailWithStatus(c, http.StatusUnauthorized, "未登录或登录已过期")
		return
	}
	if err := h.svc.Logout(c.Request.Context(), claims.UserID); err != nil {
		h.log.Error("logout 失败", zap.Error(err))
		response.FailWithStatus(c, http.StatusInternalServerError, "服务器内部错误")
		return
	}
	response.OK(c, gin.H{})
}

// writeAuthErr 将 service 层错误转为统一响应。
func (h *AuthHandler) writeAuthErr(c *gin.Context, err error) {
	if ae, ok := service.IsAppError(err); ok {
		response.Fail(c, ae.Code, ae.Code, ae.Message)
		return
	}
	if errors.Is(err, jwt.ErrTokenInvalid) || errors.Is(err, jwt.ErrTokenExpired) {
		response.Fail(c, http.StatusUnauthorized, 401, "登录已过期或 token 无效")
		return
	}
	h.log.Error("auth service error", zap.Error(err))
	response.Fail(c, http.StatusInternalServerError, 500, "服务器内部错误")
}

// toAuthUserDTO 将 User 模型转成对外的 DTO。
func toAuthUserDTO(u *model.User) dto.AuthUserDTO {
	out := dto.AuthUserDTO{
		ID:       u.ID,
		Email:    u.Email,
		Nickname: u.Nickname,
	}
	if !u.CreatedAt.IsZero() {
		out.CreatedAt = u.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	return out
}

// toAuthResultDTO 将 service.AuthResult 转成 DTO。
func toAuthResultDTO(r *service.AuthResult) dto.AuthResultDTO {
	return dto.AuthResultDTO{
		User:         toAuthUserDTO(r.User),
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		ExpiresIn:    r.ExpiresIn,
	}
}
// GetMe GET /api/v1/me
// 返回当前用户的完整资料（含 created_at）。邮箱 / 注册时间等字段只读不可改。
func (h *AuthHandler) GetMe(c *gin.Context) {
	claims, ok := userClaims(c)
	if !ok {
		return
	}
	u, err := h.svc.GetMe(c.Request.Context(), claims.UserID)
	if err != nil {
		writeErr(c, h.log, err)
		return
	}
	response.OK(c, toAuthUserDTO(u))
}

// UpdateMe PATCH /api/v1/me
// 修改昵称；邮箱 / 创建时间不可改（接口层不暴露对应字段）。
func (h *AuthHandler) UpdateMe(c *gin.Context) {
	claims, ok := userClaims(c)
	if !ok {
		return
	}
	var req dto.UpdateMeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	u, err := h.svc.UpdateMe(c.Request.Context(), claims.UserID, req.Nickname)
	if err != nil {
		writeErr(c, h.log, err)
		return
	}
	response.OK(c, toAuthUserDTO(u))
}

// ChangePassword POST /api/v1/auth/change-password（需 JWT 中间件前置）。
//
// 成功 → 更新密码 hash 并 bump token_version，使旧 refresh_token 失效；
// 旧 access_token 仍可用至自然过期，由前端在收到 200 后清凭证并跳登录页。
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	claims, ok := userClaims(c)
	if !ok {
		return
	}
	var req dto.ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.ChangePassword(c.Request.Context(), claims.UserID, service.ChangePasswordInput{
		OldPassword:     req.OldPassword,
		NewPassword:     req.NewPassword,
		ConfirmPassword: req.ConfirmPassword,
	}); err != nil {
		h.writeAuthErr(c, err)
		return
	}
	response.OK(c, gin.H{})
}
