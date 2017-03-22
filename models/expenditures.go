package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Expenditure represents a single expenditure.
type Expenditure struct {
	gorm.Model

	Amount float64   `gorm:"not null"`
	Date   time.Time `gorm:"not null"`

	Category   *Category `gorm:"ForeignKey:CategoryID"`
	CategoryID uint
}
