// Package dto - 大模型配置相关 DTO。
package dto

// CreateLLMConfigReq POST /llm-configs 请求体。
type CreateLLMConfigReq struct {
	Name      string `json:"name" binding:"required,min=1,max=20"`
	Provider  string `json:"provider" binding:"required,oneof=openai anthropic deepseek qwen zhipu moonshot custom"`
	Model     string `json:"model" binding:"required,min=1,max=100"`
	BaseURL   string `json:"base_url" binding:"required,url"`
	APIKey    string `json:"api_key" binding:"required,min=1"`
	IsEnabled *bool  `json:"is_enabled"`
	IsDefault *bool  `json:"is_default"`
}

// UpdateLLMConfigReq PUT /llm-configs/:id 请求体（全字段可选；api_key 留空=不改）。
type UpdateLLMConfigReq struct {
	Name      *string `json:"name,omitempty" binding:"omitempty,min=1,max=20"`
	Provider  *string `json:"provider,omitempty" binding:"omitempty,oneof=openai anthropic deepseek qwen zhipu moonshot custom"`
	Model     *string `json:"model,omitempty" binding:"omitempty,min=1,max=100"`
	BaseURL   *string `json:"base_url,omitempty" binding:"omitempty,url"`
	APIKey    *string `json:"api_key,omitempty" binding:"omitempty,min=1"`
	IsEnabled *bool   `json:"is_enabled"`
	IsDefault *bool   `json:"is_default"`
}

// LLMConfigDTO 对外响应（api_key_masked 永远是脱敏后的形如 `sk-fak***7890`）。
type LLMConfigDTO struct {
	ID            uint   `json:"id"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	UserID        uint   `json:"user_id"`
	Name          string `json:"name"`
	Provider      string `json:"provider"`
	ProviderLabel string `json:"provider_label"`
	Model         string `json:"model"`
	BaseURL       string `json:"base_url"`
	APIKeyMasked  string `json:"api_key_masked"`
	IsEnabled     bool   `json:"is_enabled"`
	IsDefault     bool   `json:"is_default"`
}

// TestConnectionResp 测试连接结果。
type TestConnectionResp struct {
	OK           bool   `json:"ok"`
	LatencyMs    int64  `json:"latency_ms"`
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}
// FetchModelsReq POST /llm-configs/fetch-models 请求体
type FetchModelsReq struct {
	BaseURL string `json:"base_url" binding:"required,url"`
	APIKey  string `json:"api_key" binding:"required,min=1"`
}

// FetchModelsResp 拉取模型列表响应
type FetchModelsResp struct {
	Models []string `json:"models"`
	Error  string   `json:"error,omitempty"`
}
