package service

import (
	"SService/dao"
	"SService/model"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct{}

// Register 处理用户注册逻辑
func (s *UserService) Register(nickname, email, password string) error {
	// 1. 检查邮箱是否已存在
	existingUser, err := dao.FindUserByEmail(email)
	if err == nil && existingUser != nil {
		return errors.New("邮箱已注册")
	}

	// 2. 检查昵称是否已存在（昵称用于展示，保持唯一）
	existingUser, err = dao.FindUserByNickname(nickname)
	if err == nil && existingUser != nil {
		return errors.New("昵称已存在")
	}

	// 3. 生成UUID（用户唯一标识）
	userUUID := uuid.New()

	// 4. 密码加密（重要：禁止明文存储密码）
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	// 5. 构造用户模型
	user := &model.User{
		Nickname: nickname,
		Email:    email,
		Password: string(hashedPassword), // 存储加密后的密码
		UUID:     userUUID.String(),
	}

	// 6. 调用仓库层保存用户
	if err := dao.CreateUser(user); err != nil {
		return errors.New("注册失败，请重试")
	}

	return nil
}

func (s *UserService) Login(email, password string) (*model.User, bool) {
	user, err := dao.FindUserByEmail(email)
	if err != nil {
		return nil, false
	}

	// 使用bcrypt验证密码（对比明文密码和加密后的密码）
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		// 密码不匹配
		return nil, false
	}
	return user, true
}
