package database

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// ErrItemTitleEmpty is returned when an item title is empty after trimming
// surrounding whitespace.
var ErrItemTitleEmpty = errors.New("item title must not be empty")

// UpdateItem changes an item's title and description. The title must be
// non-empty after trimming; the description is stored verbatim (an empty
// string clears it). The item must exist and belong to the caller; cross-user
// access is reported as ErrItemNotFound rather than leaking existence, mirroring
// the convention applied to every other owned update.
func UpdateItem(db *gorm.DB, userID uint, itemID uint, title, description string) (*Item, error) {
	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" {
		return nil, ErrItemTitleEmpty
	}

	var item Item
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := findItem(tx, userID, itemID, &item); err != nil {
			return err
		}
		if err := tx.Model(&item).Updates(map[string]any{
			"title":       trimmedTitle,
			"description": description,
		}).Error; err != nil {
			return fmt.Errorf("failed to update the item: %w", err)
		}
		return findItem(tx, userID, itemID, &item)
	})
	if err != nil {
		return nil, err
	}
	return &item, nil
}