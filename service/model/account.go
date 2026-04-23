package model

import "gorm.io/gorm"

// Account 资金账户（钱袋 / 独立核算单元）
// v2 去掉 type / is_virtual；账户名就是用户心智。
type Account struct {
	gorm.Model
	UserID     uint64 `json:"user_id" gorm:"not null;index:idx_account_user_archived_sort,priority:1;comment:所属用户"`
	Name       string `json:"name" gorm:"type:varchar(64);not null;comment:显示名"`
	Balance    Money  `json:"balance" gorm:"type:decimal(14,2);not null;default:0;comment:当前余额（应用层维护）"`
	Currency   string `json:"currency" gorm:"type:char(3);not null;default:'CNY';comment:货币"`
	IsArchived bool   `json:"is_archived" gorm:"not null;default:false;index:idx_account_user_archived_sort,priority:2;comment:是否已归档"`
	Sort       int    `json:"sort" gorm:"not null;default:0;index:idx_account_user_archived_sort,priority:3"`
	Note       string `json:"note" gorm:"type:varchar(255);comment:备注"`
}

// TableName 保持与文档命名一致（单数）
func (Account) TableName() string { return "account" }
