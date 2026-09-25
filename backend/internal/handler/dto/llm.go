// Package dto - LLM 调优 / 生成相关请求 / 响应 DTO。
package dto

// TuneReq POST /llm/tune 请求体（PRD P07 / F09）。
//
// 字段：
//   - prompt_id:        必填，被调优的提示词 ID（必须属于当前 user）
//   - request_text:     必填，1-2000，调优诉求
//   - use_context:      可选，是否附带分类与描述作为上下文
//   - model_config_id:  可选；不传则用 user 默认配置（若默认不存在则报错）
//   - mock:             可选，true 时跳过真实 LLM 调用，按规则拼装模拟输出（本地 / CI 用）
type TuneReq struct {
	PromptID      uint   `json:"prompt_id" binding:"required,min=1"`
	RequestText   string `json:"request_text" binding:"required,min=1,max=2000"`
	UseContext    bool   `json:"use_context"`
	ModelConfigID *uint  `json:"model_config_id" binding:"omitempty,min=1"`
	Mock          bool   `json:"mock"`
}

// GenerateReq POST /llm/generate 请求体（PRD F08 / P04 模式 B）。
//
// 字段：
//   - purpose:        必填，1-1000，用户用自然语言描述想要的提示词
//   - style:          可选，professional / friendly / concise / detailed
//   - language:       可选，zh / en / auto
//   - model_config_id: 可选；不传则用 user 默认配置
//   - mock:           可选，true 时按规则拼装模拟输出（本地 / CI 用）
type GenerateReq struct {
	Purpose       string `json:"purpose" binding:"required,min=1,max=1000"`
	Style         string `json:"style" binding:"omitempty,oneof=professional friendly concise detailed"`
	Language      string `json:"language" binding:"omitempty,oneof=zh en auto"`
	ModelConfigID *uint  `json:"model_config_id" binding:"omitempty,min=1"`
	Mock          bool   `json:"mock"`
}