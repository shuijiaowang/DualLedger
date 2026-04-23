package dao

import (
	"SService/db"
	"SService/model"
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
