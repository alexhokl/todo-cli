package database

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// UpdateItemDueDate sets an item's due date, or clears it when dueDate is nil.
func UpdateItemDueDate(db *gorm.DB, userID uint, itemID uint, dueDate *time.Time) (*Item, error) {
	var item Item
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := findItem(tx, userID, itemID, &item); err != nil {
			return err
		}
		if err := tx.Model(&item).Updates(map[string]any{"due_date": dueDate}).Error; err != nil {
			return fmt.Errorf("failed to update the due date of the item: %w", err)
		}
		return findItem(tx, userID, itemID, &item)
	})
	if err != nil {
		return nil, err
	}
	return &item, nil
}
