package dao

import (
	"SService/db"
	"SService/model"
)

func ListCategories() ([]model.CategoryEntity, error) {
	var rows []model.CategoryEntity
	err := db.DB.Order("sort asc, id asc").Find(&rows).Error
	return rows, err
}

func CategoryCodeExists(code string) (bool, error) {
	var cnt int64
	err := db.DB.Model(&model.CategoryEntity{}).Where("code = ?", code).Count(&cnt).Error
	return cnt > 0, err
}

func CategoryNameExists(name string) (bool, error) {
	var cnt int64
	err := db.DB.Model(&model.CategoryEntity{}).Where("name = ?", name).Count(&cnt).Error
	return cnt > 0, err
}

func CreateCategory(in *model.CategoryEntity) error {
	return db.DB.Create(in).Error
}

func UpdateCategory(id uint64, updates map[string]any) error {
	return db.DB.Model(&model.CategoryEntity{}).Where("id = ?", id).Updates(updates).Error
}

func DeleteCategory(id uint64) error {
	return db.DB.Delete(&model.CategoryEntity{}, id).Error
}

func GetCategoryByID(id uint64) (*model.CategoryEntity, error) {
	var row model.CategoryEntity
	if err := db.DB.First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func GetCategoryByCode(code string) (*model.CategoryEntity, error) {
	var row model.CategoryEntity
	if err := db.DB.Where("code = ?", code).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func GetCategoryByName(name string) (*model.CategoryEntity, error) {
	var row model.CategoryEntity
	if err := db.DB.Where("name = ?", name).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}
