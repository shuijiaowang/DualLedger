package model

import (
	"time"

	"gorm.io/gorm"
)

// AccrualEntry 权责真实事件条目（v2 不存 AUTO 行）
// source ∈ {MANUAL, END_SETTLE, ADJUST}
// amount 为带符号金额：正数=消耗/产出，负数=冲减。
type AccrualEntry struct {
	gorm.Model
	UserID        uint64      `json:"user_id" gorm:"not null;index:idx_accrual_user_accrue,priority:1"`
	TransactionID *uint64     `json:"transaction_id" gorm:"index;comment:关联源交易；补录可空"`
	ResourceID    *uint64     `json:"resource_id" gorm:"index:idx_accrual_resource_accrue,priority:1;comment:关联资源；一次性调整可空"`
	CategoryCode  string      `json:"category_code" gorm:"type:varchar(32);index:idx_accrual_category_accrue,priority:1;comment:分类 code（默认继承）"`
	Amount        Money       `json:"amount" gorm:"type:decimal(14,2);not null;comment:金额；允许负值（冲减）"`
	Qty           *float64    `json:"qty" gorm:"type:decimal(14,4);comment:消耗数量"`
	Unit          string      `json:"unit" gorm:"type:varchar(16)"`
	AccrueAt      time.Time   `json:"accrue_at" gorm:"not null;index:idx_accrual_user_accrue,priority:2;index:idx_accrual_resource_accrue,priority:2;index:idx_accrual_category_accrue,priority:2;comment:权责发生时间"`
	Source        string      `json:"source" gorm:"type:varchar(16);not null;comment:MANUAL/END_SETTLE/ADJUST"`
	Tags          JSONStrings `json:"tags" gorm:"type:json;comment:标签（默认继承源交易/资源）"`
	Note          string      `json:"note" gorm:"type:varchar(255)"`
}

func (AccrualEntry) TableName() string { return "accrual_entry" }
