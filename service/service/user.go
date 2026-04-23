package service

import (
	"SService/dao"
	"SService/db"
	"SService/model"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct{}

var onboardingSvc = OnboardingService{}

// Register 处理用户注册逻辑；事务内同时创建默认主账户（文档 §5.1）
func (s *UserService) Register(nickname, email, password string) error {
	existingUser, err := dao.FindUserByEmail(email)
	if err == nil && existingUser != nil && existingUser.ID != 0 {
		return errors.New("邮箱已注册")
	}

	existingUser, err = dao.FindUserByNickname(nickname)
	if err == nil && existingUser != nil && existingUser.ID != 0 {
		return errors.New("昵称已存在")
	}

	userUUID := uuid.New()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	user := &model.User{
		Nickname: nickname,
		Email:    email,
		Password: string(hashedPassword),
		UUID:     userUUID.String(),
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := dao.CreateUserTx(tx, user); err != nil {
			return errors.New("注册失败，请重试")
		}
		// 事务内为新用户建主账户；失败则整体回滚
		if err := onboardingSvc.EnsureDefaultAccount(tx, uint64(user.ID)); err != nil {
			return errors.New("初始化主账户失败：" + err.Error())
		}
		return nil
	})
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
