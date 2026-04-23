package dao

import (
	"SService/db"
	"SService/model"

	"gorm.io/gorm"
)

func FindUserByEmail(email string) (*model.User, error) {
	var user model.User
	result := db.DB.Where("email = ?", email).First(&user)
	return &user, result.Error
}

func FindUserByNickname(nickname string) (*model.User, error) {
	var user model.User
	result := db.DB.Where("nickname = ?", nickname).First(&user)
	return &user, result.Error
}

func CreateUser(user *model.User) error {
	return db.DB.Create(user).Error
}

// CreateUserTx 事务内创建（供注册时同时建主账户使用）
func CreateUserTx(tx *gorm.DB, user *model.User) error {
	return tx.Create(user).Error
}
