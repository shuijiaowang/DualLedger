package model

import (
	"time"

	"gorm.io/gorm"
)

// Resource 可摊销资源定义（文档 §3.5）
type Resource struct {
	gorm.Model
	UserID       uint64       `json:"user_id" gorm:"not null;index:idx_resource_user_status_start,priority:1"`
	Name         string       `json:"name" gorm:"type:varchar(128);not null;comment:名称"`
	CategoryCode string       `json:"category_code" gorm:"type:varchar(32);comment:分类 code"`
	Unit         string       `json:"unit" gorm:"type:varchar(16);comment:单位 天/个/次/ml"`
	TotalQty     *float64     `json:"total_qty" gorm:"type:decimal(14,4);comment:总量"`
	RemainingQty *float64     `json:"remaining_qty" gorm:"type:decimal(14,4);comment:剩余量"`
	TotalCost    Money        `json:"total_cost" gorm:"type:decimal(14,2);not null;comment:总成本"`
	AmortizeRule AmortizeRule `json:"amortize_rule" gorm:"type:json;not null;comment:摊销规则"`
	Status       string       `json:"status" gorm:"type:varchar(16);not null;default:'ACTIVE';index:idx_resource_user_status_start,priority:2"`
	PurchaseAt   time.Time    `json:"purchase_at" gorm:"not null;comment:购买时间"`
	StartUseAt   *time.Time   `json:"start_use_at" gorm:"index:idx_resource_user_status_start,priority:3;comment:开始使用时间"`
	EndAt        *time.Time   `json:"end_at" gorm:"comment:结束时间"`
	Note         string       `json:"note" gorm:"type:varchar(512)"`
	Ext          JSONMap      `json:"ext" gorm:"type:json;comment:扩展字段"`
}

func (Resource) TableName() string { return "resource" }
