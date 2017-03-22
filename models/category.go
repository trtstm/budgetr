package models

import "github.com/jinzhu/gorm"

// Category represents a single category.
type Category struct {
	gorm.Model

	Name string `gorm:"not null,unique"`
}
