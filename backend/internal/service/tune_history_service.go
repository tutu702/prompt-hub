// Package service - TuneHistoryService 调优历史业务编排。
package service

import (
	"context"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/prompthub/prompthub/internal/model"
	"github.com/prompthub/prompthub/internal/repository"
	"github.com/prompthub/prompthub/pkg/errcode"
)

// TuneHistoryService 调优历史服务。
type TuneHistoryService struct {
	histories repository.TuneHistoryRepo
	prompts   repository.PromptRepo // 用于校验 prompt 归属 + Apply 时更新 content
	log       *zap.Logger
}

// NewTuneHistoryService 构造 TuneHistoryService。
func NewTuneHistoryService(
	histories repository.TuneHistoryRepo,
	prompts repository.PromptRepo,
	log *zap.Logger,
) *TuneHistoryService {
	return &TuneHistoryService{histories: histories, prompts: prompts, log: log}
}

// TuneHistoryInput 创建入参。
type TuneHistoryInput struct {
	RequestText     string
	OriginalContent string
	TunedContent    string
	ModelConfigID   *uint
}

// TuneHistoryListResult 列表返回结果。
type TuneHistoryListResult struct {
	List  []*model.TuneHistory
	Total int64
	Page  int
	Size  int
}

// Create 创建一条历史；同时校验 prompt 归属并自增 version_no。
func (s *TuneHistoryService) Create(
	ctx context.Context,
	userID, promptID uint,
	in TuneHistoryInput,
) (*model.TuneHistory, error) {
	// 校验 prompt 存在且归属
	p, err := s.prompts.GetByID(ctx, userID, promptID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, newAppError(errcode.CodeNotFound, "提示词不存在")
	}

	// 字段长度校验
	req := strings.TrimSpace(in.RequestText)
	if req == "" || len(req) > 2000 {
		return nil, newAppError(errcode.CodeBadRequest, "调优诉求长度需为 1-2000")
	}
	if len(in.OriginalContent) == 0 || len(in.OriginalContent) > 10000 {
		return nil, newAppError(errcode.CodeBadRequest, "原文长度需为 1-10000")
	}
	if len(in.TunedContent) == 0 || len(in.TunedContent) > 10000 {
		return nil, newAppError(errcode.CodeBadRequest, "调优后内容长度需为 1-10000")
	}

	// version_no 自增：取当前 prompt 历史中最大值 + 1
	maxV, err := s.histories.MaxVersionNo(ctx, userID, promptID)
	if err != nil {
		return nil, err
	}

	h := &model.TuneHistory{
		UserID:          userID,
		PromptID:        promptID,
		VersionNo:       maxV + 1,
		RequestText:     req,
		OriginalContent: in.OriginalContent,
		TunedContent:    in.TunedContent,
		ModelConfigID:   in.ModelConfigID,
		IsApplied:       false,
	}
	if err := s.histories.Create(ctx, h); err != nil {
		return nil, err
	}
	return h, nil
}

// Get 获取单条历史；越权 404。
func (s *TuneHistoryService) Get(ctx context.Context, userID, id uint) (*model.TuneHistory, error) {
	h, err := s.histories.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if h == nil {
		return nil, newAppError(errcode.CodeNotFound, "调优历史不存在")
	}
	return h, nil
}

// List 列出某 prompt 的历史（按 version_no DESC）。
func (s *TuneHistoryService) List(
	ctx context.Context,
	userID, promptID uint,
	page, size int,
) (*TuneHistoryListResult, error) {
	// 先校验 prompt 归属（防枚举：越权访问也走 404）
	p, err := s.prompts.GetByID(ctx, userID, promptID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, newAppError(errcode.CodeNotFound, "提示词不存在")
	}

	list, total, err := s.histories.List(ctx, repository.TuneHistoryListFilter{
		UserID:   userID,
		PromptID: promptID,
		Page:     page,
		PageSize: size,
	})
	if err != nil {
		return nil, err
	}
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	return &TuneHistoryListResult{
		List:  list,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

// Apply 把历史的 tuned_content 写入 prompt.content，bump version_no，并标记历史已应用。
// 三步串行：content 更新 → version_no+1 → mark applied。失败靠前端重试（P08 接真实 LLM 时引入事务）。
func (s *TuneHistoryService) Apply(
	ctx context.Context,
	userID, historyID uint,
) (*model.Prompt, error) {
	h, err := s.Get(ctx, userID, historyID)
	if err != nil {
		return nil, err
	}

	p, err := s.prompts.GetByID(ctx, userID, h.PromptID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, newAppError(errcode.CodeNotFound, "提示词不存在")
	}

	// 1. 更新 prompt.content
	if err := s.prompts.Update(ctx, userID, h.PromptID, map[string]any{
		"content": h.TunedContent,
	}); err != nil {
		return nil, err
	}
	// 2. bump version_no（content 变化 +1，与 service.PromptService.Update 一致）
	if err := s.prompts.Update(ctx, userID, h.PromptID, map[string]any{
		"version_no": gorm.Expr("version_no + 1"),
	}); err != nil {
		return nil, err
	}
	// 3. 标记历史已应用
	if err := s.histories.MarkApplied(ctx, userID, historyID); err != nil {
		return nil, err
	}

	// 返回更新后的 prompt，便于前端同步缓存
	return s.prompts.GetByID(ctx, userID, h.PromptID)
}