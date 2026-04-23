package dao

import (
	"SService/db"
	"SService/model"
	"time"

	"gorm.io/gorm"
)

// CreateTransaction 创建交易
func CreateTransaction(tx *gorm.DB, t *model.Transaction) error {
	return conn(tx).Create(t).Error
}

// GetTransaction 查询单条
func GetTransaction(userID, id uint64) (*model.Transaction, error) {
	var t model.Transaction
	err := db.DB.Where("id = ? AND user_id = ?", id, userID).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// TransactionQuery 列表过滤参数
type TransactionQuery struct {
	UserID    uint64
	From      *time.Time
	To        *time.Time
	Types     []string
	AccountID *uint64
	Limit     int
	Offset    int
}

// ListTransactions 列表（按 occur_at 倒序）
func ListTransactions(q TransactionQuery) ([]model.Transaction, int64, error) {
	var rows []model.Transaction
	var total int64

	tx := db.DB.Model(&model.Transaction{}).Where("user_id = ?", q.UserID)
	if q.From != nil {
		tx = tx.Where("occur_at >= ?", *q.From)
	}
	if q.To != nil {
		tx = tx.Where("occur_at < ?", *q.To)
	}
	if len(q.Types) > 0 {
		tx = tx.Where("type IN ?", q.Types)
	}
	if q.AccountID != nil {
		aid := *q.AccountID
		tx = tx.Where("(account_id = ? OR to_account_id = ?)", aid, aid)
	}
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if q.Limit <= 0 {
		q.Limit = 50
	}
	if err := tx.Order("occur_at desc, id desc").
		Limit(q.Limit).Offset(q.Offset).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// DeleteTransaction 软删除 + 级联软删派生的 accrual_entry
func DeleteTransaction(userID, id uint64) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).
			Delete(&model.Transaction{}, id).Error; err != nil {
			return err
		}
		return tx.Where("user_id = ? AND transaction_id = ?", userID, id).
			Delete(&model.AccrualEntry{}).Error
	})
}

// SumBalanceFromTx 基于 transaction 汇总指定账户净流水（用于重算 balance）
// 返回有符号净值：IN 加 / OUT 减 / TRANSFER 按 account_id 与 to_account_id 分别双向影响
func SumBalanceFromTx(userID, accountID uint64) (model.Money, error) {
	type row struct {
		Type        string
		Direction   string
		Amount      string
		AccountID   uint64
		ToAccountID *uint64
	}
	var rows []row
	err := db.DB.Raw(`
		SELECT type, direction, amount, account_id, to_account_id
		FROM `+"`transaction`"+`
		WHERE user_id = ? AND deleted_at IS NULL
		  AND (account_id = ? OR to_account_id = ?)
	`, userID, accountID, accountID).Scan(&rows).Error
	if err != nil {
		return "0.00", err
	}
	net := model.NewMoney("0")
	for _, r := range rows {
		amount := model.NewMoney(r.Amount)
		if r.Type == model.TxTransfer {
			// 转账双侧：account_id 出、to_account_id 入
			if r.AccountID == accountID {
				net = net.Sub(amount)
			}
			if r.ToAccountID != nil && *r.ToAccountID == accountID {
				net = net.Add(amount)
			}
			continue
		}
		// 非转账：只看 account_id，按 direction 正负
		if r.AccountID != accountID {
			continue
		}
		switch r.Direction {
		case model.DirIn:
			net = net.Add(amount)
		case model.DirOut:
			net = net.Sub(amount)
		}
	}
	return net, nil
}
