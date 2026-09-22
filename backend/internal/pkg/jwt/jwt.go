// Package jwt 提供基于 github.com/golang-jwt/jwt/v5 的 HS256 token 签发与解析。
//
// 设计要点：
//   - Access token 短期（默认 2h），承载用户身份；
//   - Refresh token 长期（默认 7d），仅用于换取新的 access token；
//   - TokenVersion 字段预留给后续"修改密码/下线所有设备"场景，本期签发时写入但不强制校验；
//   - Parse 失败统一为 ErrTokenInvalid / ErrTokenExpired 哨兵错误，由 handler 转换为 401。
package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 哨兵错误，供业务层做边界处理。
var (
	// ErrTokenInvalid token 无法解析（格式、签名、issuer 等任一不通过）。
	ErrTokenInvalid = errors.New("token invalid")
	// ErrTokenExpired token 已过期（超过 exp + leeway）。
	ErrTokenExpired = errors.New("token expired")
)

// Claims 自定义 JWT 载荷。
type Claims struct {
	UserID       uint `json:"uid"`
	TokenVersion int  `json:"tv"`
	jwt.RegisteredClaims
}

// Manager 持有签名密钥与生命周期配置。
type Manager struct {
	secret     []byte
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
	now        func() time.Time // 可注入，便于测试
}

// NewManager 创建 JWT 管理器。secret 不能为空串。
func NewManager(secret, issuer string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{
		secret:     []byte(secret),
		issuer:     issuer,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		now:        time.Now,
	}
}

// IssueAccess 签发短期 access token。
func (m *Manager) IssueAccess(userID uint, tokenVersion int) (string, time.Time, error) {
	return m.issue(userID, tokenVersion, m.accessTTL, "access")
}

// IssueRefresh 签发长期 refresh token。
func (m *Manager) IssueRefresh(userID uint, tokenVersion int) (string, time.Time, error) {
	return m.issue(userID, tokenVersion, m.refreshTTL, "refresh")
}

func (m *Manager) issue(userID uint, tokenVersion int, ttl time.Duration, subject string) (string, time.Time, error) {
	now := m.now()
	exp := now.Add(ttl)
	claims := &Claims{
		UserID:       userID,
		TokenVersion: tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("签名失败: %w", err)
	}
	return signed, exp, nil
}

// Parse 解析并校验 token。失败返回 ErrTokenInvalid 或 ErrTokenExpired。
func (m *Manager) Parse(tokenStr string) (*Claims, error) {
	parser := jwt.NewParser(
		jwt.WithIssuer(m.issuer),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(30*time.Second),
	)
	tok, err := parser.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}
	claims, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}

// AccessTTL 返回 access token 有效期（用于响应 expires_in 字段）。
func (m *Manager) AccessTTL() time.Duration {
	return m.accessTTL
}