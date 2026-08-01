package database

import (
	"gorm.io/gorm"
)

// AutoMigrate creates or updates the database schema for all known models.
func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&User{},
		&TailscaleAddress{},
		&List{},
		&Label{},
		&Effort{},
		&Blocker{},
		&Item{},
	); err != nil {
		return err
	}

	return nil
}
