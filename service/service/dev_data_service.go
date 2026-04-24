package service

import (
	"SService/db"
	"SService/model"
	"fmt"

	"gorm.io/gorm"
)

type DevDataService struct{}

func (s *DevDataService) ResetForUser(userID uint64) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&model.AccrualEntry{}).Error; err != nil {
			return fmt.Errorf("清理 accrual_entry 失败: %w", err)
		}
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&model.Transaction{}).Error; err != nil {
			return fmt.Errorf("清理 transaction 失败: %w", err)
		}
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&model.Resource{}).Error; err != nil {
			return fmt.Errorf("清理 resource 失败: %w", err)
		}
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&model.Account{}).Error; err != nil {
			return fmt.Errorf("清理 account 失败: %w", err)
		}
		return nil
	})
}
