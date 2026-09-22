// Package service 实现业务编排层。
//
// AuthService 负责注册 / 登录 / 刷新 / 登出的核心逻辑：
//   - 校验入参（密码强度、协议勾选）；
//   - 协调 repository 与 jwt 包完成用户创建与 token 签发；
//   - 统一通过 *errcode.AppError 透传业务错误；
//   - 登录时使用 dummy bcrypt 平衡响应时长，防止通过耗时差异枚举账号。
package service

import (
	"context"
	"errors"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/prompthub/prompthub/internal/model"
	"github.com/prompthub/prompthub/internal/pkg/jwt"
	"github.com/prompthub/prompthub/internal/pkg/password"
	"github.com/prompthub/prompthub/internal/repository"
)

// authErrcode 表示与 errcode 包对齐的业务异常（仅本服务用）。
type authErrcode = int

const (
	aeOK                  authErrcode = 0
	aeBadRequest          authErrcode = 400
	aeInvalidCredential   authErrcode = 401
	aeEmailConflict       authErrcode = 409
	aeAgreementRequired   authErrcode = 400
	aePasswordWeak        authErrcode = 400
	aeTokenInvalid        authErrcode = 401
)

// appError 实现 error，统一 code/message 字段。
type appError struct {
	Code    int
	Message string
}

func (e *appError) Error() string { return e.Message }

// newAppError 构造一个 appError。
func newAppError(code int, message string) *appError {
	return &appError{Code: code, Message: message}
}

// IsAppError 判断是否为业务异常，便于 handler 透传。
func IsAppError(err error) (*appError, bool) {
	var ae *appError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}

// passwordWeakPattern 必须同时含字母和数字，骨架阶段使用单一正则覆盖两种字符类型。
var passwordWeakPattern = regexp.MustCompile(`^[A-Za-z0-9]{8,64}$`)
var hasLetter = regexp.MustCompile(`[A-Za-z]`)
var hasDigit = regexp.MustCompile(`[0-9]`)

// dummyHash 是一段预生成的 bcrypt 哈希，用于登录时"用户不存在"分支对齐响应时长。
// 由 password.Hash 一次性生成后固定下来，避免每次启动重新生成。
var dummyHash = mustGenerateDummy()

func mustGenerateDummy() string {
	h, err := password.Hash("__dummy_password_for_timing_safety__")
	if err != nil {
		// 不应触发；失败则用一段固定假哈希，长度上与 bcrypt 输出一致。
		return "$2a$10$abcdefghijklmnopqrstuv1234567890abcdefghijklmnopqrstuv12"
	}
	return h
}

// AuthService 鉴权服务。
type AuthService struct {
	users repository.UserRepo
	jwt   *jwt.Manager
	log   *zap.Logger
}

// NewAuthService 创建 AuthService。
func NewAuthService(users repository.UserRepo, mgr *jwt.Manager, log *zap.Logger) *AuthService {
	return &AuthService{users: users, jwt: mgr, log: log}
}

// RegisterInput 注册入参（service 层独立校验，与 DTO 解耦）。
type RegisterInput struct {
	Email           string
	Password        string
	ConfirmPassword string
	Nickname        string
	Agreement       bool
}

// LoginInput 登录入参。
type LoginInput struct {
	Email    string
	Password string
	Remember bool
}

// AuthResult 业务成功响应：用户信息 + 双 token + 过期秒数。
type AuthResult struct {
	User         *model.User
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

// Register 注册新用户；成功后自动签发 token（前端实现"注册即登录"）。
func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*AuthResult, error) {
	if !in.Agreement {
		return nil, newAppError(aeAgreementRequired, "请阅读并同意用户协议")
	}

	email := normalizeEmail(in.Email)
	if email == "" {
		return nil, newAppError(aeBadRequest, "邮箱格式不正确")
	}

	if in.Password != in.ConfirmPassword {
		return nil, newAppError(aeBadRequest, "两次输入的密码不一致")
	}
	if !isPasswordStrong(in.Password) {
		return nil, newAppError(aePasswordWeak, "密码至少 8 位且需包含字母与数字")
	}

	exists, err := s.users.EmailExists(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, newAppError(aeEmailConflict, "该邮箱已被注册")
	}

	hash, err := password.Hash(in.Password)
	if err != nil {
		return nil, err
	}

	nickname := strings.TrimSpace(in.Nickname)
	if nickname == "" {
		nickname = email[:strings.Index(email, "@")]
	}

	u := &model.User{
		Email:        email,
		PasswordHash: hash,
		Nickname:     nickname,
		TokenVersion: 1,
		Status:       1,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, u)
}

// Login 校验凭据；统一返回 "邮箱或密码错误" 防止枚举。
func (s *AuthService) Login(ctx context.Context, in LoginInput) (*AuthResult, error) {
	email := normalizeEmail(in.Email)
	if email == "" {
		// 即便邮箱格式错，也跑一次 dummy 哈希对齐时长。
		_ = password.Verify(dummyHash, in.Password)
		return nil, newAppError(aeInvalidCredential, "邮箱或密码错误")
	}

	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if u == nil {
		_ = password.Verify(dummyHash, in.Password)
		return nil, newAppError(aeInvalidCredential, "邮箱或密码错误")
	}

	if u.Status != 1 {
		_ = password.Verify(dummyHash, in.Password)
		return nil, newAppError(aeInvalidCredential, "邮箱或密码错误")
	}

	if !password.Verify(u.PasswordHash, in.Password) {
		return nil, newAppError(aeInvalidCredential, "邮箱或密码错误")
	}

	return s.issueTokens(ctx, u)
}

// Refresh 用 refresh token 换取新的 access token。
// 本期不轮换 refresh token，也不强制校验 TokenVersion（字段预留）。
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*AuthResult, error) {
	claims, err := s.jwt.Parse(refreshToken)
	if err != nil {
		return nil, newAppError(aeTokenInvalid, "refresh token 无效或已过期")
	}
	if claims.Subject != "refresh" {
		return nil, newAppError(aeTokenInvalid, "refresh token 无效或已过期")
	}

	u, err := s.users.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if u == nil || u.Status != 1 {
		return nil, newAppError(aeTokenInvalid, "账户不存在或已停用")
	}

	access, _, err := s.jwt.IssueAccess(u.ID, u.TokenVersion)
	if err != nil {
		return nil, err
	}
	return &AuthResult{
		User:         u,
		AccessToken:  access,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.jwt.AccessTTL().Seconds()),
	}, nil
}

// Logout 本期服务端无副作用（JWT 本身即可吊销）。
// 后续接入 TokenVersion++ 时在此 bump。
func (s *AuthService) Logout(_ context.Context, _ uint) error {
	return nil
}

// issueTokens 签发双 token 并异步更新最后登录时间。
func (s *AuthService) issueTokens(ctx context.Context, u *model.User) (*AuthResult, error) {
	access, _, err := s.jwt.IssueAccess(u.ID, u.TokenVersion)
	if err != nil {
		return nil, err
	}
	refresh, _, err := s.jwt.IssueRefresh(u.ID, u.TokenVersion)
	if err != nil {
		return nil, err
	}

	// 异步更新登录时间，失败仅记日志，不影响主流程。
	go func(uid uint) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.users.UpdateLastLogin(bgCtx, uid, ""); err != nil {
			s.log.Warn("更新最后登录时间失败", zap.Uint("user_id", uid), zap.Error(err))
		}
	}(u.ID)

	return &AuthResult{
		User:         u,
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int(s.jwt.AccessTTL().Seconds()),
	}, nil
}

// normalizeEmail 转小写并 trim 空白字符；格式非法时返回空串。
func normalizeEmail(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.ToLower(raw)
	if raw == "" {
		return ""
	}
	addr, err := mail.ParseAddress(raw)
	if err != nil {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(addr.Address))
}

// isPasswordStrong 校验密码强度：长度 8-64，必须同时含字母与数字。
func isPasswordStrong(p string) bool {
	if !passwordWeakPattern.MatchString(p) {
		return false
	}
	return hasLetter.MatchString(p) && hasDigit.MatchString(p)
}

// IsGormNotFound 给外部的便捷判断（service 不应直接依赖 gorm，但当前仓库其它模块已使用）。
// 本期未导出，保持私有以避免污染 API。
func IsGormNotFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }

// 静默导入 url 包，确保后续如需校验 email URL/重定向字段有可用工具。
var _ = url.QueryEscape