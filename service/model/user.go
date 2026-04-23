package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Nickname string `json:"nickname" gorm:"type:varchar(50);uniqueIndex;not null;comment:昵称"`
	Email    string `json:"email" gorm:"type:varchar(100);uniqueIndex;not null;comment:邮箱"`
	Password string `json:"-" gorm:"type:varchar(100);not null;comment:加密密码"`
	UUID     string `json:"uuid" gorm:"type:char(36);uniqueIndex;not null;comment:用户唯一标识UUID"`
}
