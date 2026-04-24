package service

import (
	"SService/dao"
	"SService/model"
	"fmt"
	"strings"
)

type CategoryService struct{}

func (s *CategoryService) List() ([]model.CategoryEntity, error) {
	return dao.ListCategories()
}

func (s *CategoryService) Create(in model.CategoryEntity) (*model.CategoryEntity, error) {
	if in.Name == "" || in.Kind == "" {
		return nil, fmt.Errorf("name/kind 必填")
	}
	if in.Source == "" {
		in.Source = "user"
	}
	in.Code = strings.TrimSpace(in.Code)
	in.Name = strings.TrimSpace(in.Name)
	if in.Code == "" {
		var err error
		in.Code, err = buildCategoryCodeFromName(in.Name)
		if err != nil {
			return nil, err
		}
	}
	nameExists, err := dao.CategoryNameExists(in.Name)
	if err != nil {
		return nil, err
	}
	if nameExists {
		return nil, fmt.Errorf("名称已存在: %s", in.Name)
	}
	ok, err := dao.CategoryCodeExists(in.Code)
	if err != nil {
		return nil, err
	}
	if ok {
		return nil, fmt.Errorf("code 已存在: %s", in.Code)
	}
	if err := dao.CreateCategory(&in); err != nil {
		return nil, err
	}
	return &in, nil
}

func (s *CategoryService) Update(id uint64, updates map[string]any) error {
	if _, hasCode := updates["code"]; hasCode {
		return fmt.Errorf("不允许修改 code")
	}
	return dao.UpdateCategory(id, updates)
}

func (s *CategoryService) Delete(id uint64) error {
	return dao.DeleteCategory(id)
}

func buildCategoryCodeFromName(name string) (string, error) {
	base := strings.ToLower(strings.TrimSpace(name))
	base = strings.ReplaceAll(base, " ", "_")
	base = strings.ReplaceAll(base, ".", "_")
	base = strings.ReplaceAll(base, "-", "_")
	base = strings.ReplaceAll(base, "/", "_")
	if base == "" {
		return "", fmt.Errorf("名称不能为空")
	}
	candidate := base
	for i := 1; i <= 1000; i++ {
		ok, err := dao.CategoryCodeExists(candidate)
		if err != nil {
			return "", err
		}
		if !ok {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s_%d", base, i)
	}
	return "", fmt.Errorf("自动生成分类标识失败，请重试")
}
