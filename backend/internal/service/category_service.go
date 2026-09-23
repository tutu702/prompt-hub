// Package service - CategoryService 分类业务编排。
//
// 业务规则：
//   - 同一用户下分类名不可重复；
//   - 删除前必须确保无提示词引用，否则返回 409；
//   - color 可空；非空时必须符合 #RRGGBB 格式。
package service

import (
	"context"
	"regexp"
	"strings"

	"go.uber.org/zap"

	"github.com/prompthub/prompthub/internal/model"
	"github.com/prompthub/prompthub/internal/repository"
	"github.com/prompthub/prompthub/pkg/errcode"
)

// CategoryService 分类服务。
type CategoryService struct {
	repo    repository.CategoryRepo
	log     *zap.Logger
	promptN func(ctx context.Context, userID, categoryID uint) (int64, error)
}

// NewCategoryService 构造 CategoryService。
// promptN 是"统计某分类下提示词数量"的回调，由 server 层注入 PromptRepo.CountByCategory，
// 避免 service 层反向依赖 prompt-repo。
func NewCategoryService(repo repository.CategoryRepo, promptN func(ctx context.Context, userID, categoryID uint) (int64, error), log *zap.Logger) *CategoryService {
	return &CategoryService{repo: repo, log: log, promptN: promptN}
}

// CategoryInput 分类输入（指针表示可空）。
type CategoryInput struct {
	Name  string
	Color string
	Sort  *int
}

// normalizeColor 校验并规范 color；空字符串保留（视为使用默认）。
func normalizeColor(raw string) (string, error) {
	c := strings.TrimSpace(raw)
	if c == "" {
		return "", nil
	}
	if !colorRegexp.MatchString(c) {
		return "", newAppError(errcode.CodeBadRequest, "颜色格式不正确，应为 #RRGGBB")
	}
	return strings.ToUpper(c), nil
}

var colorRegexp = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

// Create 创建分类。
func (s *CategoryService) Create(ctx context.Context, userID uint, in CategoryInput) (*model.Category, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" || len(name) > 20 {
		return nil, newAppError(errcode.CodeBadRequest, "分类名称长度需为 1-20")
	}
	color, err := normalizeColor(in.Color)
	if err != nil {
		return nil, err
	}

	// 同名预检
	existing, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, c := range existing {
		if c.Name == name {
			return nil, newAppError(errcode.CodeConflict, "已存在同名分类")
		}
	}

	sort := 0
	if in.Sort != nil {
		sort = *in.Sort
	}

	c := &model.Category{
		UserID: userID,
		Name:   name,
		Color:  color,
		Sort:   sort,
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Get 获取分类。
func (s *CategoryService) Get(ctx context.Context, userID, id uint) (*model.Category, error) {
	c, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, newAppError(errcode.CodeNotFound, "分类不存在")
	}
	return c, nil
}

// List 列出当前用户的所有分类。
func (s *CategoryService) List(ctx context.Context, userID uint) ([]*model.Category, error) {
	return s.repo.List(ctx, userID)
}

// Update 更新分类（指针为 nil 表示不改）。
func (s *CategoryService) Update(ctx context.Context, userID, id uint, in CategoryInput) (*model.Category, error) {
	fields := map[string]any{}
	if in.Name != "" {
		name := strings.TrimSpace(in.Name)
		if len(name) > 20 {
			return nil, newAppError(errcode.CodeBadRequest, "分类名称长度需为 1-20")
		}
		fields["name"] = name
	}
	if in.Color != "" {
		color, err := normalizeColor(in.Color)
		if err != nil {
			return nil, err
		}
		fields["color"] = color
	} else if in.Color != "" && len(in.Color) == 0 {
		// 显式清空（保留扩展）
	}
	if in.Sort != nil {
		fields["sort"] = *in.Sort
	}
	if len(fields) == 0 {
		return s.Get(ctx, userID, id)
	}

	// 校验存在性
	if _, err := s.Get(ctx, userID, id); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, userID, id, fields); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, userID, id)
}

// Delete 删除分类；若仍有提示词引用，返回 409。
func (s *CategoryService) Delete(ctx context.Context, userID, id uint) error {
	if _, err := s.Get(ctx, userID, id); err != nil {
		return err
	}
	if s.promptN != nil {
		n, err := s.promptN(ctx, userID, id)
		if err != nil {
			return err
		}
		if n > 0 {
			return newAppError(errcode.CodeConflict, "请先将提示词移至其他分类")
		}
	}
	return s.repo.Delete(ctx, userID, id)
}