package model

import "gorm.io/gorm"

// CategoryEntity 分类表（支持系统分类 + 用户维护）
type CategoryEntity struct {
	gorm.Model
	Code       string `json:"code" gorm:"type:varchar(32);not null;uniqueIndex"`
	ParentCode string `json:"parent_code,omitempty" gorm:"type:varchar(32);index:idx_category_parent_sort,priority:1"`
	Name       string `json:"name" gorm:"type:varchar(64);not null"`
	Kind       string `json:"kind" gorm:"type:varchar(16);not null;comment:INCOME/EXPENSE/TRANSFER/OTHER"`
	Icon       string `json:"icon,omitempty" gorm:"type:varchar(64)"`
	Sort       int    `json:"sort" gorm:"not null;default:0;index:idx_category_parent_sort,priority:2"`
	Source     string `json:"source" gorm:"type:varchar(16);not null;default:system;comment:system/user"`
}

func (CategoryEntity) TableName() string { return "category" }
