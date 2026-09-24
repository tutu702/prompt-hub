// Package dto 集中放置 HTTP 入参 / 出参的数据传输对象。
//
// 校验规则：尽量用 gin 的 binding tag 描述；只有自定义规则（如密码正则）下沉到 service 层。
package dto

// RegisterReq 注册请求体。
type RegisterReq struct {
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=8,max=64"`
	ConfirmPassword string `json:"confirmPassword" binding:"required,eqfield=Password"`
	Nickname        string `json:"nickname" binding:"omitempty,max=32"`
	Agreement       bool   `json:"agreement" binding:"required"`
}

// LoginReq 登录请求体。
type LoginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=64"`
	Remember bool   `json:"remember"`
}

// RefreshReq 刷新 access token 请求体。
type RefreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// AuthUserDTO 用户在鉴权响应中的脱敏形式（不暴露 password_hash / token_version）。
type AuthUserDTO struct {
	ID        uint   `json:"id"`
	Email     string `json:"email"`
	Nickname  string `json:"nickname"`
	CreatedAt string `json:"created_at,omitempty"`
}

// AuthResultDTO 鉴权成功响应统一格式。
type AuthResultDTO struct {
	User         AuthUserDTO `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int         `json:"expires_in"`
}
// UpdateMeReq PATCH /me 请求体（只允许改昵称）。
type UpdateMeReq struct {
	Nickname string `json:"nickname" binding:"required,min=1,max=20"`
}

// ChangePasswordReq 修改密码请求体（PRD P10 /settings/password）。
//
// 服务端校验顺序：
//   1. 三字段均必填（binding 已保证）；
//   2. confirmPassword == newPassword；
//   3. newPassword 强度（8-64，含字母+数字）；
//   4. oldPassword 与现有 hash 匹配。
type ChangePasswordReq struct {
	OldPassword     string `json:"oldPassword" binding:"required,min=1,max=64"`
	NewPassword     string `json:"newPassword" binding:"required,min=8,max=64"`
	ConfirmPassword string `json:"confirmPassword" binding:"required,min=8,max=64"`
}
