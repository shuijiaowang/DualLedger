package dao

import (
	"SService/db"
	"SService/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func CreateResource(tx *gorm.DB, r *model.Resource) error {
	return conn(tx).Create(r).Error
}

func GetResource(userID, id uint64) (*model.Resource, error) {
	var r model.Resource
	err := db.DB.Where("id = ? AND user_id = ?", id, userID).First(&r).Error
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func GetResourceForUpdate(tx *gorm.DB, userID, id uint64) (*model.Resource, error) {
	var r model.Resource
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND user_id = ?", id, userID).
		First(&r).Error
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func ListResources(userID uint64, statuses []string) ([]model.Resource, error) {
	var rs []model.Resource
	q := db.DB.Where("user_id = ?", userID)
	if len(statuses) > 0 {
		q = q.Where("status IN ?", statuses)
	}
	err := q.Order("id desc").Find(&rs).Error
	return rs, err
}

func UpdateResource(tx *gorm.DB, userID, id uint64, updates map[string]any) error {
	return conn(tx).Model(&model.Resource{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(updates).Error
}
