package model

import "gorm.io/gorm"

// Example 占位业务表：演示 AutoMigrate 与 JWT 保护下的接口，可按需替换字段。
type Example struct {
	gorm.Model
}
