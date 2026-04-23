package dao

import (
	"SService/db"
	"SService/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CreateAccount 创建账户（调用方负责填 UserID/Name/初始 Balance）
// 支持传入 tx 以便在业务事务里使用
func CreateAccount(tx *gorm.DB, acc *model.Account) error {
	return conn(tx).Create(acc).Error
}

// ListAccounts 列出某用户的全部账户（按排序），含已归档
func ListAccounts(userID uint64, includeArchived bool) ([]model.Account, error) {
	var accs []model.Account
	q := db.DB.Where("user_id = ?", userID)
	if !includeArchived {
		q = q.Where("is_archived = ?", false)
	}
	err := q.Order("sort asc, id asc").Find(&accs).Error
	return accs, err
}

// GetAccount 按 id + userID 查询（权限隔离）
func GetAccount(userID, id uint64) (*model.Account, error) {
	var acc model.Account
	err := db.DB.Where("id = ? AND user_id = ?", id, userID).First(&acc).Error
	if err != nil {
		return nil, err
	}
	return &acc, nil
}

// GetAccountForUpdate 事务内锁行（FOR UPDATE），维护账户余额时使用
func GetAccountForUpdate(tx *gorm.DB, userID, id uint64) (*model.Account, error) {
	var acc model.Account
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND user_id = ?", id, userID).
		First(&acc).Error
	if err != nil {
		return nil, err
	}
	return &acc, nil
}

// UpdateAccount 更新指定字段
func UpdateAccount(tx *gorm.DB, userID, id uint64, updates map[string]any) error {
	return conn(tx).Model(&model.Account{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(updates).Error
}

// DeleteAccount 软删除
func DeleteAccount(userID, id uint64) error {
	return db.DB.Where("user_id = ?", userID).Delete(&model.Account{}, id).Error
}

// AdjustBalance 原子地增量调整余额（delta 为带符号 Money）
func AdjustBalance(tx *gorm.DB, accountID uint64, delta model.Money) error {
	// 直接走 SQL 保证原子；使用 DECIMAL 运算
	return conn(tx).Exec(
		"UPDATE `account` SET balance = balance + ?, updated_at = NOW() WHERE id = ? AND deleted_at IS NULL",
		delta.String(), accountID,
	).Error
}

// conn 允许 DAO 复用外层事务
func conn(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return db.DB
}
