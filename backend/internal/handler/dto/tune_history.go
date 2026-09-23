// Package dto - 调优历史相关 DTO。
package dto

// CreateTuneHistoryReq POST /prompts/:id/tune-histories 请求体。
type CreateTuneHistoryReq struct {
	RequestText     string `json:"request_text" binding:"required,min=1,max=2000"`
	OriginalContent string `json:"original_content" binding:"required,min=1,max=10000"`
	TunedContent    string `json:"tuned_content" binding:"required,min=1,max=10000"`
	ModelConfigID   *uint  `json:"model_config_id" binding:"omitempty"`
}

// ListTuneHistoriesReq GET /prompts/:id/tune-histories 查询参数。
type ListTuneHistoriesReq struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// TuneHistoryDTO 调优历史对外响应。
type TuneHistoryDTO struct {
	ID              uint    `json:"id"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
	UserID          uint    `json:"user_id"`
	PromptID        uint    `json:"prompt_id"`
	VersionNo       int     `json:"version_no"`
	RequestText     string  `json:"request_text"`
	OriginalContent string  `json:"original_content"`
	TunedContent    string  `json:"tuned_content"`
	ModelConfigID   *uint   `json:"model_config_id,omitempty"`
	IsApplied       bool    `json:"is_applied"`
}