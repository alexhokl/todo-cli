package database

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

var (
	// ErrBlockerNotFound is returned when a blocker does not exist.
	ErrBlockerNotFound = errors.New("blocker not found")
	// ErrBlockerDescriptionEmpty is returned when a blocker description is empty
	// after trimming surrounding whitespace.
	ErrBlockerDescriptionEmpty = errors.New("blocker description must not be empty")
)

// ListBlockers returns every blocker attached to the given item, ordered by
// identifier (creation order). The item must exist and belong to the caller;
// cross-user access is reported as not found rather than leaking existence.
func ListBlockers(db *gorm.DB, userID uint, itemID uint) ([]Blocker, error) {
	var item Item
	if err := findItem(db, userID, itemID, &item); err != nil {
		return nil, err
	}

	var blockers []Blocker
	if err := db.Where("item_id = ?", itemID).Where("user_id = ?", userID).
		Order("id ASC").Find(&blockers).Error; err != nil {
		return nil, fmt.Errorf("failed to list the blockers: %w", err)
	}

	return blockers, nil
}

// CreateBlocker attaches a new blocker to an item. The item must exist and
// belong to the caller; the description must be non-empty after trimming.
func CreateBlocker(db *gorm.DB, userID uint, itemID uint, description string) (*Blocker, error) {
	trimmed := strings.TrimSpace(description)
	if trimmed == "" {
		return nil, ErrBlockerDescriptionEmpty
	}

	var item Item
	var blocker Blocker
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := findItem(tx, userID, itemID, &item); err != nil {
			return err
		}

		blocker = Blocker{
			Description: trimmed,
			ItemID:      itemID,
			UserID:      userID,
		}
		if err := tx.Create(&blocker).Error; err != nil {
			return fmt.Errorf("failed to create the blocker: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &blocker, nil
}

// UpdateBlocker changes the description of an existing blocker. The blocker
// must exist and belong to the caller; the description must be non-empty after
// trimming.
func UpdateBlocker(db *gorm.DB, userID uint, blockerID uint, description string) (*Blocker, error) {
	trimmed := strings.TrimSpace(description)
	if trimmed == "" {
		return nil, ErrBlockerDescriptionEmpty
	}

	var blocker Blocker
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := findBlocker(tx, userID, blockerID, &blocker); err != nil {
			return err
		}

		// Model(&Blocker{}) with an explicit Where avoids the GORM stale-struct
		// gotcha where Model(&instance) uses the instance's field values.
		if err := tx.Model(&Blocker{}).Where("id = ?", blockerID).
			UpdateColumn("description", trimmed).Error; err != nil {
			return fmt.Errorf("failed to update the blocker: %w", err)
		}

		// Re-read into a fresh struct so preloaded fields do not retain stale
		// values from before the update.
		blocker = Blocker{}
		return findBlocker(tx, userID, blockerID, &blocker)
	})
	if err != nil {
		return nil, err
	}

	return &blocker, nil
}

// DeleteBlocker removes a blocker. The blocker must exist and belong to the
// caller; cross-user access is reported as not found.
func DeleteBlocker(db *gorm.DB, userID uint, blockerID uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var blocker Blocker
		if err := findBlocker(tx, userID, blockerID, &blocker); err != nil {
			return err
		}

		if err := tx.Delete(&blocker).Error; err != nil {
			return fmt.Errorf("failed to delete the blocker: %w", err)
		}

		return nil
	})
}

// findBlocker loads a blocker by identifier, translating a missing row into
// ErrBlockerNotFound. The query is scoped to the given user so cross-user
// access is reported as not found rather than leaking existence.
func findBlocker(tx *gorm.DB, userID uint, id uint, blocker *Blocker) error {
	err := tx.Where("user_id = ?", userID).First(blocker, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("%w: %d", ErrBlockerNotFound, id)
	}
	if err != nil {
		return fmt.Errorf("failed to query the blocker: %w", err)
	}

	return nil
}