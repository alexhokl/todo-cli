package database

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

var (
	// ErrItemLinkToSelf is returned when an item is linked to itself.
	ErrItemLinkToSelf = errors.New("an item cannot be linked to itself")
)

// itemLinks is the join table model for the symmetric self-referential
// many-to-many between items. GORM auto-creates the item_links table from the
// LinkedItems tag on Item, but this struct is used for explicit two-row
// symmetric insert/delete with idempotency, which GORM's default association
// helpers do not support for a self-referential relationship.
type itemLinks struct {
	ItemID        uint `gorm:"primaryKey"`
	LinkedItemID  uint `gorm:"primaryKey"`
}

// TableName pins the join table name so GORM does not pluralise the struct
// name (item_links is already the desired snake_case form).
func (itemLinks) TableName() string { return "item_links" }

// UpdateItemLinks attaches and detaches links between an item and other items.
// The relationship is symmetric: linking A to B also links B to A, stored as
// two join rows so GORM's Preload reads correctly from either side. Self-links
// are rejected, and target items must exist and belong to the caller. Adding
// an already-present link and removing one that is absent are both no-ops.
func UpdateItemLinks(db *gorm.DB, userID uint, itemID uint, addIDs, removeIDs []uint) (*Item, error) {
	var item Item
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := findItem(tx, userID, itemID, &item); err != nil {
			return err
		}

		// Reject self-links and resolve every target against the caller's own
		// items so cross-user access is reported as not found rather than
		// leaking existence.
		for _, id := range addIDs {
			if id == itemID {
				return ErrItemLinkToSelf
			}
			if err := findItem(tx, userID, id, &Item{}); err != nil {
				return err
			}
		}
		for _, id := range removeIDs {
			if id == itemID {
				return ErrItemLinkToSelf
			}
			if err := findItem(tx, userID, id, &Item{}); err != nil {
				return err
			}
		}

		// Removal happens first so that passing the same id to both flags
		// leaves the link attached rather than depending on statement order.
		// Both directions of the symmetric pair are removed.
		for _, id := range removeIDs {
			if err := tx.Where("item_id = ? AND linked_item_id = ?", itemID, id).
				Delete(&itemLinks{}).Error; err != nil {
				return fmt.Errorf("failed to remove the link: %w", err)
			}
			if err := tx.Where("item_id = ? AND linked_item_id = ?", id, itemID).
				Delete(&itemLinks{}).Error; err != nil {
				return fmt.Errorf("failed to remove the reverse link: %w", err)
			}
		}

		for _, id := range addIDs {
			// Idempotent: skip if the link already exists in either direction.
			if linkExists(tx, itemID, id) {
				continue
			}
			if err := tx.Create(&itemLinks{ItemID: itemID, LinkedItemID: id}).Error; err != nil {
				return fmt.Errorf("failed to add the link: %w", err)
			}
			if err := tx.Create(&itemLinks{ItemID: id, LinkedItemID: itemID}).Error; err != nil {
				return fmt.Errorf("failed to add the reverse link: %w", err)
			}
		}

		// Reload with the preloaded associations the caller expects.
		item = Item{}
		return findItem(tx, userID, itemID, &item)
	})
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// linkExists reports whether a symmetric link between a and b is already
// present in either direction.
func linkExists(tx *gorm.DB, a, b uint) bool {
	var count int64
	tx.Model(&itemLinks{}).
		Where("item_id = ? AND linked_item_id = ?", a, b).
		Or("item_id = ? AND linked_item_id = ?", b, a).
		Count(&count)
	return count > 0
}