package service

import (
	"SService/dao"
	"SService/db"
	"SService/model"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type LedgerService struct{}

// TxInput 创建 transaction 的入参
type TxInput struct {
	UserID       uint64
	Type         string
	Direction    string // 留空则按 type 推导
	OccurAt      time.Time
	Amount       model.Money
	AccountID    uint64
	ToAccountID  *uint64 // 仅 TRANSFER
	CategoryCode string
	ResourceID   *uint64
	Counterparty string
	Title        string
	Note         string
	Ext          model.JSONMap
}

// Create 创建 transaction，事务内同步账户余额
func (s *LedgerService) Create(in TxInput) (*model.Transaction, error) {
	if err := validateTxInput(&in); err != nil {
		return nil, err
	}

	var result *model.Transaction
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		// 1) 账户权限校验
		if _, err := dao.GetAccount(in.UserID, in.AccountID); err != nil {
			return fmt.Errorf("账户不存在或无权访问: %w", err)
		}
		if in.ToAccountID != nil {
			if _, err := dao.GetAccount(in.UserID, *in.ToAccountID); err != nil {
				return fmt.Errorf("目标账户不存在或无权访问: %w", err)
			}
		}

		// 2) 组装
		t := &model.Transaction{
			UserID:       in.UserID,
			Type:         in.Type,
			Direction:    in.Direction,
			OccurAt:      in.OccurAt,
			Amount:       in.Amount,
			AccountID:    in.AccountID,
			ToAccountID:  in.ToAccountID,
			CategoryCode: in.CategoryCode,
			ResourceID:   in.ResourceID,
			Counterparty: in.Counterparty,
			Title:        in.Title,
			Note:         in.Note,
			Ext:          in.Ext,
		}
		if err := dao.CreateTransaction(tx, t); err != nil {
			return err
		}

		// 3) 同步账户余额（ADJUST 不动现金，TRANSFER 双侧）
		if err := applyBalance(tx, t); err != nil {
			return err
		}

		result = t
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// RebuildBalance 按 transaction 全量重算账户余额（修数据）
func (s *LedgerService) RebuildBalance(userID, accountID uint64) (model.Money, error) {
	acc, err := dao.GetAccount(userID, accountID)
	if err != nil {
		return "0.00", fmt.Errorf("账户不存在: %w", err)
	}
	net, err := dao.SumBalanceFromTx(userID, accountID)
	if err != nil {
		return "0.00", err
	}
	if err := dao.UpdateAccount(nil, userID, uint64(acc.ID), map[string]any{"balance": net.String()}); err != nil {
		return "0.00", err
	}
	return net, nil
}

// validateTxInput 业务不变式 §15.1 ～ §15.7
func validateTxInput(in *TxInput) error {
	if !model.IsValidTxType(in.Type) {
		return fmt.Errorf("非法 transaction.type: %s", in.Type)
	}
	if in.Direction == "" {
		in.Direction = model.DefaultDirectionFor(in.Type)
	}
	if !model.IsValidDirection(in.Direction) {
		return fmt.Errorf("非法 direction: %s", in.Direction)
	}
	// amount > 0 恒成立
	if in.Amount.Cmp(model.NewMoney("0")) <= 0 {
		return errors.New("amount 必须 > 0；方向由 type+direction 表达")
	}
	if in.OccurAt.IsZero() {
		in.OccurAt = time.Now()
	}
	if in.AccountID == 0 {
		return errors.New("account_id 必填")
	}
	if in.CategoryCode != "" && !model.CategoryExists(in.CategoryCode) {
		return fmt.Errorf("未知 category_code: %s", in.CategoryCode)
	}
	// TRANSFER：必须有 to_account_id 且与 account_id 不同
	if in.Type == model.TxTransfer {
		if in.ToAccountID == nil {
			return errors.New("TRANSFER 必须提供 to_account_id")
		}
		if *in.ToAccountID == in.AccountID {
			return errors.New("TRANSFER 的出入账户不能相同")
		}
		if in.Direction != model.DirBoth {
			return errors.New("TRANSFER 的 direction 必须为 BOTH")
		}
	} else {
		if in.ToAccountID != nil {
			return fmt.Errorf("type=%s 不允许提供 to_account_id", in.Type)
		}
	}
	return nil
}

// applyBalance 按 type + direction 更新账户余额（事务内调用）
func applyBalance(tx *gorm.DB, t *model.Transaction) error {
	switch t.Type {
	case model.TxTransfer:
		if err := dao.AdjustBalance(tx, t.AccountID, t.Amount.Negate()); err != nil {
			return err
		}
		if t.ToAccountID != nil {
			if err := dao.AdjustBalance(tx, *t.ToAccountID, t.Amount); err != nil {
				return err
			}
		}
	case model.TxAdjust:
		// ADJUST 不动现金；仅作语义事件（文档 §4.7）
		return nil
	default:
		// INCOME / EXPENSE / LOAN / DEPOSIT / REFUND：按 direction
		switch t.Direction {
		case model.DirIn:
			return dao.AdjustBalance(tx, t.AccountID, t.Amount)
		case model.DirOut:
			return dao.AdjustBalance(tx, t.AccountID, t.Amount.Negate())
		}
	}
	return nil
}
