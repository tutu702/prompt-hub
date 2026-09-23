// Package dto - 分类相关 DTO。
package dto

// CreateCategoryReq POST /categories 请求体。
type CreateCategoryReq struct {
	Name  string `json:"name" binding:"required,min=1,max=20"`
	Color string `json:"color" binding:"omitempty,max=7"`
	Sort  *int   `json:"sort" binding:"omitempty"`
}

// UpdateCategoryReq PUT /categories/:id 请求体。
type UpdateCategoryReq struct {
	Name  *string `json:"name,omitempty" binding:"omitempty,min=1,max=20"`
	Color *string `json:"color,omitempty" binding:"omitempty,max=7"`
	Sort  *int    `json:"sort,omitempty"`
}

// CategoryDTO 分类对外响应。
type CategoryDTO struct {
	ID           uint   `json:"id"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	UserID       uint   `json:"user_id"`
	Name         string `json:"name"`
	Color        string `json:"color"`
	Sort         int    `json:"sort"`
	PromptCount  int64  `json:"prompt_count,omitempty"`
}