package controllers

import (
	"github.com/jinzhu/gorm"
	"github.com/trtstm/budgetr/db"
)

func categoryStatsQuery() *gorm.DB {
	q := db.DB.Table("expenditures")
	q = q.Joins("LEFT JOIN categories ON expenditures.category_id = categories.id")
	q = q.Group("expenditures.category_id")
	q = q.Where("expenditures.deleted_at IS NULL")
	q = q.Select("categories.id as id, categories.name AS name, SUM(expenditures.amount) as total")

	return q
}
