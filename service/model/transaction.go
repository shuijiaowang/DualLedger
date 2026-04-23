package model

import (
	"time"

	"gorm.io/gorm"
)

// Transaction 交易主单（v2 吸收 cash_entry 职责）
// 约定：amount 恒 > 0；方向由 type + direction 表达，不存负数。
type Transaction struct {
	gorm.Model
	UserID       uint64    `json:"user_id" gorm:"not null;index:idx_tx_user_occur,priority:1;index:idx_tx_user_type_occur,priority:1"`
	Type         string    `json:"type" gorm:"type:varchar(16);not null;index:idx_tx_user_type_occur,priority:2;comment:INCOME/EXPENSE/TRANSFER/LOAN/DEPOSIT/REFUND/ADJUST"`
	Direction    string    `json:"direction" gorm:"type:varchar(4);not null;comment:IN/OUT/BOTH"`
	OccurAt      time.Time `json:"occur_at" gorm:"not null;index:idx_tx_user_occur,priority:2;index:idx_tx_user_type_occur,priority:3;index:idx_tx_account_occur,priority:2;comment:业务发生时间"`
	Amount       Money     `json:"amount" gorm:"type:decimal(14,2);not null;comment:正数金额"`
	AccountID    uint64    `json:"account_id" gorm:"not null;index:idx_tx_account_occur,priority:1;comment:主账户（TRANSFER 为出方）"`
	ToAccountID  *uint64   `json:"to_account_id" gorm:"index:idx_tx_to_account_occur,priority:1;comment:TRANSFER 入方"`
	CategoryCode string    `json:"category_code" gorm:"type:varchar(32);comment:分类 code"`
	ResourceID   *uint64   `json:"resource_id" gorm:"index;comment:关联资源"`
	Counterparty string    `json:"counterparty" gorm:"type:varchar(64);comment:对手方（LOAN/DEPOSIT/REFUND 使用）"`
	Title        string    `json:"title" gorm:"type:varchar(128)"`
	Note         string    `json:"note" gorm:"type:varchar(1024)"`
	Ext          JSONMap   `json:"ext" gorm:"type:json;comment:扩展字段"`
}

func (Transaction) TableName() string { return "transaction" }
