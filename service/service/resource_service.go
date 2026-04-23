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

type ResourceService struct{}

// CreateResourceInput 同时写 resource + 关联 transaction 的入参
// 买电动牙刷、一箱苹果、工资发薪都走这个。
type CreateResourceInput struct {
	UserID       uint64
	Name         string
	CategoryCode string
	Unit         string
	TotalQty     *float64
	TotalCost    model.Money
	AmortizeRule model.AmortizeRule
	PurchaseAt   time.Time
	StartUseAt   *time.Time
	Note         string
	Ext          model.JSONMap

	// 关联交易（可选：传 AccountID 则同时创建 EXPENSE / INCOME）
	AccountID uint64
	TxType    string // EXPENSE / INCOME，留空不创建交易
	TxTitle   string
}

// CreateResult 创建结果
type CreateResult struct {
	Resource    *model.Resource    `json:"resource"`
	Transaction *model.Transaction `json:"transaction,omitempty"`
}

// Create 创建资源（可选同时扣款）
func (s *ResourceService) Create(in CreateResourceInput) (*CreateResult, error) {
	if in.Name == "" {
		return nil, errors.New("resource.name 必填")
	}
	if in.TotalCost.Cmp(model.NewMoney("0")) <= 0 {
		return nil, errors.New("total_cost 必须 > 0")
	}
	if err := in.AmortizeRule.Validate(); err != nil {
		return nil, err
	}
	if in.CategoryCode != "" && !model.CategoryExists(in.CategoryCode) {
		return nil, fmt.Errorf("未知 category_code: %s", in.CategoryCode)
	}
	if in.PurchaseAt.IsZero() {
		in.PurchaseAt = time.Now()
	}
	if in.TxType != "" && in.TxType != model.TxExpense && in.TxType != model.TxIncome {
		return nil, errors.New("resource 关联交易只能是 EXPENSE 或 INCOME")
	}

	var res *model.Resource
	var tx *model.Transaction
	err := db.DB.Transaction(func(dbTx *gorm.DB) error {
		remaining := in.TotalQty
		r := &model.Resource{
			UserID:       in.UserID,
			Name:         in.Name,
			CategoryCode: in.CategoryCode,
			Unit:         in.Unit,
			TotalQty:     in.TotalQty,
			RemainingQty: remaining,
			TotalCost:    in.TotalCost,
			AmortizeRule: in.AmortizeRule,
			Status:       model.ResStatusActive,
			PurchaseAt:   in.PurchaseAt,
			StartUseAt:   in.StartUseAt,
			Note:         in.Note,
			Ext:          in.Ext,
		}
		if err := dao.CreateResource(dbTx, r); err != nil {
			return err
		}
		res = r

		if in.TxType != "" {
			if in.AccountID == 0 {
				return errors.New("关联交易必须指定 account_id")
			}
			txTitle := in.TxTitle
			if txTitle == "" {
				txTitle = in.Name
			}
			direction := model.DefaultDirectionFor(in.TxType)
			t := &model.Transaction{
				UserID:       in.UserID,
				Type:         in.TxType,
				Direction:    direction,
				OccurAt:      in.PurchaseAt,
				Amount:       in.TotalCost,
				AccountID:    in.AccountID,
				CategoryCode: in.CategoryCode,
				ResourceID:   pu64(uint64(r.ID)),
				Title:        txTitle,
				Ext:          in.Ext,
			}
			if err := dao.CreateTransaction(dbTx, t); err != nil {
				return err
			}
			if err := applyBalance(dbTx, t); err != nil {
				return err
			}
			tx = t
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &CreateResult{Resource: res, Transaction: tx}, nil
}

// PunchInput BY_COUNT 打卡
type PunchInput struct {
	UserID     uint64
	ResourceID uint64
	Qty        float64
	AccrueAt   time.Time
	Tags       []string // 覆盖默认继承；nil 表示继承 resource.ext.tags
	Note       string
	MarkEnded  bool // 这次打卡后是否直接结束（已消耗完）
}

// Punch BY_COUNT 打卡；按 剩余成本/剩余数量*qty 计算金额，保证最终闭合。
func (s *ResourceService) Punch(in PunchInput) (*model.AccrualEntry, error) {
	if in.Qty <= 0 {
		return nil, errors.New("qty 必须 > 0")
	}
	if in.AccrueAt.IsZero() {
		in.AccrueAt = time.Now()
	}

	var entry *model.AccrualEntry
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		r, err := dao.GetResourceForUpdate(tx, in.UserID, in.ResourceID)
		if err != nil {
			return fmt.Errorf("资源不存在: %w", err)
		}
		if r.AmortizeRule.Type != model.AmortizeByCount {
			return fmt.Errorf("只有 BY_COUNT 资源支持打卡，当前 type=%s", r.AmortizeRule.Type)
		}
		if r.Status != model.ResStatusActive {
			return fmt.Errorf("资源已不是 ACTIVE，当前状态=%s", r.Status)
		}
		if r.RemainingQty == nil || *r.RemainingQty <= 0 {
			return errors.New("资源已无剩余数量")
		}
		if in.Qty > *r.RemainingQty+1e-9 {
			return fmt.Errorf("本次 qty=%v 超过剩余 %v", in.Qty, *r.RemainingQty)
		}

		// 累计真实事件 + 剩余数量、剩余成本
		sum, err := dao.SumAccrualForResource(in.UserID, uint64(r.ID))
		if err != nil {
			return err
		}
		remainingCost := r.TotalCost.Sub(sum)
		// amount = remainingCost / remainingQty * qty
		// 为简单起见用浮点算（金额两位保留，1e-4 量级完全够用）
		var amount model.Money
		if *r.RemainingQty > 0 {
			unit := parseFloat(remainingCost.String()) / *r.RemainingQty
			amt := unit * in.Qty
			amount = model.NewMoney(fmt.Sprintf("%.2f", amt))
		} else {
			amount = "0.00"
		}

		tags := in.Tags
		if tags == nil {
			tags = tagsFromExt(r.Ext)
		}
		e := &model.AccrualEntry{
			UserID:       in.UserID,
			ResourceID:   pu64(uint64(r.ID)),
			CategoryCode: r.CategoryCode,
			Amount:       amount,
			Qty:          pf(in.Qty),
			Unit:         r.Unit,
			AccrueAt:     in.AccrueAt,
			Source:       model.AccrualManual,
			Tags:         model.JSONStrings(tags),
			Note:         in.Note,
		}
		if err := dao.CreateAccrualEntry(tx, e); err != nil {
			return err
		}
		entry = e

		newRemaining := *r.RemainingQty - in.Qty
		updates := map[string]any{"remaining_qty": newRemaining}
		if in.MarkEnded || newRemaining <= 1e-9 {
			updates["status"] = model.ResStatusEnded
			now := time.Now()
			updates["end_at"] = now
		}
		return dao.UpdateResource(tx, in.UserID, uint64(r.ID), updates)
	})
	if err != nil {
		return nil, err
	}
	return entry, nil
}

// EndResourceInput 结束资源（ENDED / DISCARDED / RETURNED）
type EndResourceInput struct {
	UserID      uint64
	ResourceID  uint64
	Status      string // ResStatusEnded / Discarded / Returned
	AccrueAt    time.Time
	Note        string
	WriteSettle bool // 是否写 END_SETTLE 条目把剩余成本一次性摊入（默认 true for DISCARDED/ENDED，false for RETURNED-策略A）
}

// End 结束资源；根据状态和 WriteSettle 写 accrual_entry
func (s *ResourceService) End(in EndResourceInput) (*model.Resource, *model.AccrualEntry, error) {
	if !model.IsValidResourceStatus(in.Status) || in.Status == model.ResStatusActive {
		return nil, nil, fmt.Errorf("非法终态: %s", in.Status)
	}
	if in.AccrueAt.IsZero() {
		in.AccrueAt = time.Now()
	}

	var res *model.Resource
	var entry *model.AccrualEntry
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		r, err := dao.GetResourceForUpdate(tx, in.UserID, in.ResourceID)
		if err != nil {
			return err
		}
		if r.Status != model.ResStatusActive {
			return fmt.Errorf("资源已处于终态: %s", r.Status)
		}

		if in.WriteSettle {
			// 已累计真实事件 + 动态虚拟（本次不算进去，保持与 DB 一致）
			sum, err := dao.SumAccrualForResource(in.UserID, uint64(r.ID))
			if err != nil {
				return err
			}
			remaining := r.TotalCost.Sub(sum)
			// 对 FIXED_PERIOD / DYNAMIC_BY_DAY 规则：动态部分已在视图呈现，
			// 进终态后视图把 end_at 之前的动态行留档、之后不再产出。
			// 为了"真实事件总和闭合"，此处仅把"未动态化且未真实化"的剩余计入 END_SETTLE。
			// MVP 对 FIXED_PERIOD/DYNAMIC_BY_DAY 默认不写 END_SETTLE，避免与动态行重复；
			// 对 BY_COUNT 默认写剩余成本为损失。
			writeSettle := false
			if r.AmortizeRule.Type == model.AmortizeByCount && remaining.Cmp(model.NewMoney("0")) > 0 {
				writeSettle = true
			}
			if writeSettle {
				e := &model.AccrualEntry{
					UserID:       in.UserID,
					ResourceID:   pu64(uint64(r.ID)),
					CategoryCode: r.CategoryCode,
					Amount:       remaining,
					Qty:          r.RemainingQty,
					Unit:         r.Unit,
					AccrueAt:     in.AccrueAt,
					Source:       model.AccrualEndSettle,
					Tags:         model.JSONStrings(append(tagsFromExt(r.Ext), "损失")),
					Note:         in.Note,
				}
				if err := dao.CreateAccrualEntry(tx, e); err != nil {
					return err
				}
				entry = e
			}
		}

		zero := 0.0
		updates := map[string]any{
			"status":        in.Status,
			"end_at":        in.AccrueAt,
			"remaining_qty": &zero,
		}
		if err := dao.UpdateResource(tx, in.UserID, uint64(r.ID), updates); err != nil {
			return err
		}
		// 回查
		res, err = dao.GetResource(in.UserID, uint64(r.ID))
		return err
	})
	if err != nil {
		return nil, nil, err
	}
	return res, entry, nil
}

func parseFloat(s string) float64 {
	var f float64
	_, _ = fmt.Sscanf(s, "%f", &f)
	return f
}

func pf(v float64) *float64 { return &v }

func pu64(v uint64) *uint64 { return &v }
