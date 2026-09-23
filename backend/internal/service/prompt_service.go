// Package service - PromptService 提示词业务编排。
//
// 业务规则：
//   - 所有读写强制 userID；越权 → 404（统一"提示词不存在"）；
//   - title 1-100、description 0-200、content 1-10000；
//   - categoryID 非 nil 时必须存在且属于当前用户；
//   - tags 存为 JSON 字符串（[]string ↔ string）；
//   - Update 字段语义：`*T` 指针，nil=不改；非 nil=覆盖；
//   - 列表分页归一化：page≥1、page_size 20≤x≤100、默认 sort=updated_desc、默认 q_in=all。
package service

import (
	"context"
	"encoding/json"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/prompthub/prompthub/internal/model"
	"github.com/prompthub/prompthub/internal/repository"
	"github.com/prompthub/prompthub/pkg/errcode"
)

// PromptService 提示词服务。
type PromptService struct {
	repo    repository.PromptRepo
	catRepo repository.CategoryRepo
	log     *zap.Logger
}

// NewPromptService 构造 PromptService。
func NewPromptService(repo repository.PromptRepo, catRepo repository.CategoryRepo, log *zap.Logger) *PromptService {
	return &PromptService{repo: repo, catRepo: catRepo, log: log}
}

// PromptInput 创建/全量更新入参。
type PromptInput struct {
	Title       string
	Description string
	Content     string
	CategoryID  *uint
	Tags        []string
}

// PromptUpdateInput 增量更新入参（指针=可空；nil=不改）。
type PromptUpdateInput struct {
	Title       *string
	Description *string
	Content     *string
	CategoryID  *uint
	Tags        *[]string
	IsFavorite  *bool
}

// PromptListQuery 列表查询参数。
type PromptListQuery struct {
	Keyword    string
	QIn        string
	CategoryID *uint
	IsFavorite *bool
	Sort       string
	Page       int
	PageSize   int
}

// PromptListResult 列表返回结果。
type PromptListResult struct {
	List  []*model.Prompt
	Total int64
	Page  int
	Size  int
}

// Create 创建提示词。
func (s *PromptService) Create(ctx context.Context, userID uint, in PromptInput) (*model.Prompt, error) {
	if err := validateTitleDescContent(in.Title, in.Description, in.Content); err != nil {
		return nil, err
	}
	if err := s.validateCategory(ctx, userID, in.CategoryID); err != nil {
		return nil, err
	}

	p := &model.Prompt{
		UserID:      userID,
		CategoryID:  in.CategoryID,
		Title:       strings.TrimSpace(in.Title),
		Description: strings.TrimSpace(in.Description),
		Content:     in.Content,
		IsFavorite:  false,
		Tags:        marshalTags(in.Tags),
		VersionNo:   1,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// Get 获取提示词。
func (s *PromptService) Get(ctx context.Context, userID, id uint) (*model.Prompt, error) {
	p, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, newAppError(errcode.CodeNotFound, "提示词不存在")
	}
	return p, nil
}

// List 列表查询。
func (s *PromptService) List(ctx context.Context, userID uint, q PromptListQuery) (*PromptListResult, error) {
	page := q.Page
	if page <= 0 {
		page = 1
	}
	size := q.PageSize
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	sort := q.Sort
	if sort == "" {
		sort = "updated_desc"
	}
	qIn := q.QIn
	if qIn == "" {
		qIn = "all"
	}

	list, total, err := s.repo.List(ctx, repository.PromptListFilter{
		UserID:     userID,
		Keyword:    strings.TrimSpace(q.Keyword),
		QIn:        qIn,
		CategoryID: q.CategoryID,
		IsFavorite: q.IsFavorite,
		Sort:       sort,
		Page:       page,
		PageSize:   size,
	})
	if err != nil {
		return nil, err
	}
	return &PromptListResult{
		List:  list,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

// Update 更新提示词（增量）。
func (s *PromptService) Update(ctx context.Context, userID, id uint, in PromptUpdateInput) (*model.Prompt, error) {
	// 校验存在
	if _, err := s.Get(ctx, userID, id); err != nil {
		return nil, err
	}

	fields := map[string]any{}
	bumpVersion := false

	if in.Title != nil {
		title := strings.TrimSpace(*in.Title)
		if title == "" || len(title) > 100 {
			return nil, newAppError(errcode.CodeBadRequest, "标题长度需为 1-100")
		}
		fields["title"] = title
	}
	if in.Description != nil {
		desc := *in.Description
		if len(desc) > 200 {
			return nil, newAppError(errcode.CodeBadRequest, "描述不超过 200 字符")
		}
		fields["description"] = desc
	}
	if in.Content != nil {
		c := *in.Content
		if c == "" || len(c) > 10000 {
			return nil, newAppError(errcode.CodeBadRequest, "内容长度需为 1-10000")
		}
		fields["content"] = c
		bumpVersion = true
	}
	if in.CategoryID != nil {
		if err := s.validateCategory(ctx, userID, in.CategoryID); err != nil {
			return nil, err
		}
		fields["category_id"] = *in.CategoryID
	}
	if in.Tags != nil {
		fields["tags"] = marshalTags(*in.Tags)
	}
	if in.IsFavorite != nil {
		fields["is_favorite"] = *in.IsFavorite
	}

	if len(fields) == 0 {
		return s.Get(ctx, userID, id)
	}
	if bumpVersion {
		// 让 GORM 用 expression 自增；这里直接 SQL 自增
		if err := s.repo.Update(ctx, userID, id, fields); err != nil {
			return nil, err
		}
		// 单独 increment version_no
		if err := s.repo.Update(ctx, userID, id, map[string]any{
			"version_no": gorm.Expr("version_no + 1"),
		}); err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.Update(ctx, userID, id, fields); err != nil {
			return nil, err
		}
	}
	return s.repo.GetByID(ctx, userID, id)
}

// Delete 删除（软删）。
func (s *PromptService) Delete(ctx context.Context, userID, id uint) error {
	if _, err := s.Get(ctx, userID, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, userID, id)
}

// BatchDelete 批量删除。
func (s *PromptService) BatchDelete(ctx context.Context, userID uint, ids []uint) (int64, error) {
	if len(ids) == 0 {
		return 0, newAppError(errcode.CodeBadRequest, "请选择要删除的提示词")
	}
	return s.repo.BatchDelete(ctx, userID, ids)
}

// ToggleFavorite 切换收藏。
func (s *PromptService) ToggleFavorite(ctx context.Context, userID, id uint, isFavorite bool) (*model.Prompt, error) {
	if _, err := s.Get(ctx, userID, id); err != nil {
		return nil, err
	}
	if err := s.repo.ToggleFavorite(ctx, userID, id, isFavorite); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, userID, id)
}

// --- 内部辅助 ---

// validateTitleDescContent 校验 title/description/content 长度。
func validateTitleDescContent(title, desc, content string) error {
	title = strings.TrimSpace(title)
	if title == "" || len(title) > 100 {
		return newAppError(errcode.CodeBadRequest, "标题长度需为 1-100")
	}
	if len(desc) > 200 {
		return newAppError(errcode.CodeBadRequest, "描述不超过 200 字符")
	}
	if content == "" || len(content) > 10000 {
		return newAppError(errcode.CodeBadRequest, "内容长度需为 1-10000")
	}
	return nil
}

// validateCategory 校验分类归属。
func (s *PromptService) validateCategory(ctx context.Context, userID uint, cid *uint) error {
	if cid == nil {
		return nil
	}
	c, err := s.catRepo.GetByID(ctx, userID, *cid)
	if err != nil {
		return err
	}
	if c == nil {
		return newAppError(errcode.CodeBadRequest, "分类不存在")
	}
	return nil
}

// marshalTags []string → JSON 字符串；空切片输出 "[]"。
func marshalTags(tags []string) string {
	if tags == nil {
		return "[]"
	}
	b, err := json.Marshal(tags)
	if err != nil {
		return "[]"
	}
	return string(b)
}

// UnmarshalTags 把 model.Prompt.Tags 还原为 []string；解析失败返回空切片。
func UnmarshalTags(raw string) []string {
	if raw == "" {
		return []string{}
	}
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return []string{}
	}
	if out == nil {
		return []string{}
	}
	return out
}