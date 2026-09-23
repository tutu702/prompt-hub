// Package dto - 提示词相关 DTO。
package dto

// ListPromptsReq GET /prompts 列表查询参数。
type ListPromptsReq struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	PageSize   int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Keyword    string `form:"keyword"`
	QIn        string `form:"q_in" binding:"omitempty,oneof=title all"`
	CategoryID *uint  `form:"category_id" binding:"omitempty"`
	IsFavorite *bool  `form:"is_favorite" binding:"omitempty"`
	Sort       string `form:"sort" binding:"omitempty,oneof=updated_desc created_desc title_asc"`
}

// CreatePromptReq POST /prompts 请求体。
// CategoryID 为指针类型，允许传 null（不分类）/ 数字 / 省略（不分类）。
type CreatePromptReq struct {
	Title       string   `json:"title" binding:"required,min=1,max=100"`
	Description string   `json:"description" binding:"omitempty,max=200"`
	Content     string   `json:"content" binding:"required,min=1,max=10000"`
	CategoryID  *uint    `json:"category_id" binding:"omitempty"`
	Tags        []string `json:"tags" binding:"omitempty,dive,max=32"`
}

// UpdatePromptReq PUT /prompts/:id 请求体（全字段可选）。
type UpdatePromptReq struct {
	Title         *string   `json:"title,omitempty" binding:"omitempty,min=1,max=100"`
	Description   *string   `json:"description,omitempty" binding:"omitempty,max=200"`
	Content       *string   `json:"content,omitempty" binding:"omitempty,min=1,max=10000"`
	CategoryID    *uint     `json:"category_id,omitempty"`
	Tags          *[]string `json:"tags,omitempty"`
	IsFavorite    *bool     `json:"is_favorite,omitempty"`
}

// BatchDeleteReq POST /prompts/batch-delete 请求体。
type BatchDeleteReq struct {
	IDs []uint `json:"ids" binding:"required,min=1,dive,gt=0"`
}

// ToggleFavoriteReq POST /prompts/:id/favorite 请求体。
type ToggleFavoriteReq struct {
	IsFavorite bool `json:"is_favorite"`
}

// PromptDTO 提示词对外响应。
type PromptDTO struct {
	ID          uint     `json:"id"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
	UserID      uint     `json:"user_id"`
	CategoryID  *uint    `json:"category_id,omitempty"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	IsFavorite  bool     `json:"is_favorite"`
	Tags        []string `json:"tags"`
	VersionNo   int      `json:"version_no"`
}