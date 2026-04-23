package dao

import (
	"SService/db"
	"SService/model"
	"time"

	"gorm.io/gorm"
)

func CreateAccrualEntry(tx *gorm.DB, e *model.AccrualEntry) error {
	return conn(tx).Create(e).Error
}

func DeleteAccrualEntry(userID, id uint64) error {
	return db.DB.Where("user_id = ?", userID).Delete(&model.AccrualEntry{}, id).Error
}

// AccrualQuery 权责事件查询参数
type AccrualQuery struct {
	UserID       uint64
	From         *time.Time
	To           *time.Time
	ResourceID   *uint64
	CategoryCode string
	Tag          string // JSON_CONTAINS 单标签
	Limit        int
	Offset       int
}

// ListAccrualEntries 真实事件列表
func ListAccrualEntries(q AccrualQuery) ([]model.AccrualEntry, error) {
	var rows []model.AccrualEntry
	tx := db.DB.Where("user_id = ?", q.UserID)
	if q.From != nil {
		tx = tx.Where("accrue_at >= ?", *q.From)
	}
	if q.To != nil {
		tx = tx.Where("accrue_at < ?", *q.To)
	}
	if q.ResourceID != nil {
		tx = tx.Where("resource_id = ?", *q.ResourceID)
	}
	if q.CategoryCode != "" {
		tx = tx.Where("category_code = ?", q.CategoryCode)
	}
	if q.Tag != "" {
		// MySQL JSON 函数；注意参数需要 JSON 引号
		tx = tx.Where("JSON_CONTAINS(tags, JSON_QUOTE(?))", q.Tag)
	}
	if q.Limit <= 0 {
		q.Limit = 200
	}
	err := tx.Order("accrue_at asc, id asc").
		Limit(q.Limit).Offset(q.Offset).
		Find(&rows).Error
	return rows, err
}

// SumAccrualForResource 资源累计真实事件金额（用于 END_SETTLE 结算时计算剩余成本）
func SumAccrualForResource(userID, resourceID uint64) (model.Money, error) {
	type row struct{ Amount string }
	var r row
	err := db.DB.Raw(`
		SELECT COALESCE(SUM(amount), 0) AS amount FROM accrual_entry
		WHERE user_id = ? AND resource_id = ? AND deleted_at IS NULL
	`, userID, resourceID).Scan(&r).Error
	if err != nil {
		return "0.00", err
	}
	return model.NewMoney(r.Amount), nil
}
