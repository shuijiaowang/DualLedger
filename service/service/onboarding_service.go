package service

import (
	"SService/dao"
	"SService/model"

	"gorm.io/gorm"
)

type OnboardingService struct{}

// EnsureDefaultAccount 为新用户创建主账户；幂等（已存在任何账户则跳过）
func (s *OnboardingService) EnsureDefaultAccount(tx *gorm.DB, userID uint64) error {
	existing, err := dao.ListAccounts(userID, true)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil
	}
	acc := &model.Account{
		UserID:     userID,
		Name:       "主账户",
		Balance:    model.NewMoney("0"),
		Currency:   "CNY",
		IsArchived: false,
		Sort:       0,
	}
	return dao.CreateAccount(tx, acc)
}
