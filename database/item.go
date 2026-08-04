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

// ErrItemNotUntriaged is returned when a delete is attempted on an item that is
// not in the untriaged state (done or already carrying a priority).
var ErrItemNotUntriaged = errors.New("item is not untriaged")

// ErrItemHasLinks is returned when a delete is attempted on an item that still
// has linked items. Linked items must be detached before the item is deleted.
var ErrItemHasLinks = errors.New("item has linked items")

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

// DeleteItem removes an untriaged item. Only items that are not done and carry
// no priority may be deleted; items with linked items are rejected. Attached
// blockers and comments are removed in the same operation, and the symmetric
// item_links join rows are cleaned up. The item must exist and belong to the
// caller; cross-user access is reported as ErrItemNotFound rather than leaking
// existence.
func DeleteItem(db *gorm.DB, userID uint, itemID uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var item Item
		if err := findItem(tx, userID, itemID, &item); err != nil {
			return err
		}

		if item.Done || item.Priority != nil {
			return fmt.Errorf("%w: %d", ErrItemNotUntriaged, itemID)
		}
		if len(item.LinkedItems) > 0 {
			return fmt.Errorf("%w: %d", ErrItemHasLinks, itemID)
		}

		// Remove the attached blockers and comments before deleting the item
		// so the children do not outlive their parent.
		if err := tx.Where("item_id = ? AND user_id = ?", itemID, userID).
			Delete(&Blocker{}).Error; err != nil {
			return fmt.Errorf("failed to delete the blockers: %w", err)
		}
		if err := tx.Where("item_id = ? AND user_id = ?", itemID, userID).
			Delete(&Comment{}).Error; err != nil {
			return fmt.Errorf("failed to delete the comments: %w", err)
		}

		// Clean up the symmetric item_links join rows pointing at this item in
		// either direction so no dangling back-links remain on the other side.
		if err := tx.Where("item_id = ? OR linked_item_id = ?", itemID, itemID).
			Delete(&itemLinks{}).Error; err != nil {
			return fmt.Errorf("failed to delete the item links: %w", err)
		}

		if err := tx.Delete(&item).Error; err != nil {
			return fmt.Errorf("failed to delete the item: %w", err)
		}

		return nil
	})
}