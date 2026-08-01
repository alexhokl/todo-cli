package database

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

var (
	// ErrEffortNotFound is returned when an effort does not exist.
	ErrEffortNotFound = errors.New("effort not found")
	// ErrEffortExists is returned when an effort name is already taken.
	ErrEffortExists = errors.New("effort already exists")
	// ErrEffortInUse is returned when deleting an effort that is still attached
	// to at least one item.
	ErrEffortInUse = errors.New("effort is still attached to items")
	// ErrEffortNameEmpty is returned when an effort name normalises to nothing.
	ErrEffortNameEmpty = errors.New("effort name must not be empty")
)

// NormaliseEffortName reduces an effort name to its canonical form. It is the
// single definition of effort identity: any two names with the same normalised
// form refer to the same effort.
func NormaliseEffortName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// ListEfforts returns every effort owned by the user, ordered by name.
func ListEfforts(db *gorm.DB, userID uint) ([]Effort, error) {
	var efforts []Effort
	if err := db.Where("user_id = ?", userID).Order("name ASC").Find(&efforts).Error; err != nil {
		return nil, fmt.Errorf("failed to list the efforts: %w", err)
	}

	return efforts, nil
}

// CreateEffort creates an effort explicitly. Unlike the assignment path it
// does not fall back to returning an existing effort, so that an accidental
// duplicate is reported rather than silently accepted.
func CreateEffort(db *gorm.DB, userID uint, name string) (*Effort, error) {
	normalised := NormaliseEffortName(name)
	if normalised == "" {
		return nil, ErrEffortNameEmpty
	}

	var effort Effort
	err := db.Transaction(func(tx *gorm.DB) error {
		switch err := tx.Where("name = ?", normalised).Where("user_id = ?", userID).First(&effort).Error; {
		case err == nil:
			return fmt.Errorf("%w: %s", ErrEffortExists, normalised)
		case !errors.Is(err, gorm.ErrRecordNotFound):
			return fmt.Errorf("failed to query the effort: %w", err)
		}

		effort = Effort{Name: normalised, UserID: userID}
		if err := tx.Create(&effort).Error; err != nil {
			return fmt.Errorf("failed to create the effort: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &effort, nil
}

// RenameEffort changes the name of an existing effort. Renaming an effort to
// the name it already has is a no-op rather than a conflict.
func RenameEffort(db *gorm.DB, userID uint, id uint, name string) (*Effort, error) {
	normalised := NormaliseEffortName(name)
	if normalised == "" {
		return nil, ErrEffortNameEmpty
	}

	var effort Effort
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := findEffort(tx, userID, id, &effort); err != nil {
			return err
		}
		if effort.Name == normalised {
			return nil
		}

		var existing Effort
		switch err := tx.Where("name = ?", normalised).Where("user_id = ?", userID).First(&existing).Error; {
		case err == nil:
			return fmt.Errorf("%w: %s", ErrEffortExists, normalised)
		case !errors.Is(err, gorm.ErrRecordNotFound):
			return fmt.Errorf("failed to query the effort: %w", err)
		}

		if err := tx.Model(&effort).Update("name", normalised).Error; err != nil {
			return fmt.Errorf("failed to rename the effort: %w", err)
		}

		return findEffort(tx, userID, id, &effort)
	})
	if err != nil {
		return nil, err
	}

	return &effort, nil
}

// DeleteEffort removes an effort that is no longer attached to any item. Efforts
// in use are reported rather than silently detached, so that deleting an effort
// can never quietly clear the effort of an item. Unlike labels there is no join
// table to sweep, since an item references an effort directly via a foreign key.
func DeleteEffort(db *gorm.DB, userID uint, id uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var effort Effort
		if err := findEffort(tx, userID, id, &effort); err != nil {
			return err
		}

		var count int64
		if err := tx.
			Model(&Item{}).
			Where("effort_id = ?", id).
			Where("user_id = ?", userID).
			Count(&count).Error; err != nil {
			return fmt.Errorf("failed to count the items using the effort: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("%w: %s is attached to %d item(s)", ErrEffortInUse, effort.Name, count)
		}

		if err := tx.Delete(&effort).Error; err != nil {
			return fmt.Errorf("failed to delete the effort: %w", err)
		}

		return nil
	})
}

// FindEffortByName resolves a normalised name to an existing effort owned by
// the user. It does not create the effort if it is missing; the caller must
// create it explicitly via CreateEffort first. An empty name normalises to an
// empty string and yields ErrEffortNameEmpty rather than a lookup.
func FindEffortByName(db *gorm.DB, userID uint, name string) (*Effort, error) {
	normalised := NormaliseEffortName(name)
	if normalised == "" {
		return nil, ErrEffortNameEmpty
	}

	var effort Effort
	err := db.Where("name = ?", normalised).Where("user_id = ?", userID).First(&effort).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("%w: %s", ErrEffortNotFound, normalised)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query the effort: %w", err)
	}

	return &effort, nil
}

// SetItemEffort attaches an effort to an item by name, or clears it when the
// name is empty. The effort must already exist; unknown names are reported
// rather than being created on the fly, so triage of the effort catalog stays
// an explicit step.
func SetItemEffort(db *gorm.DB, userID uint, itemID uint, name string) (*Item, error) {
	var item Item
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := findItem(tx, userID, itemID, &item); err != nil {
			return err
		}

		normalised := NormaliseEffortName(name)
		if normalised == "" {
			if err := tx.Model(&Item{}).Where("id = ?", item.ID).UpdateColumn("effort_id", nil).Error; err != nil {
				return fmt.Errorf("failed to clear the effort: %w", err)
			}
			// A fresh struct is used for the re-read so that the preloaded
			// belongs-to association does not retain the stale EffortID from
			// before the update (GORM does not zero struct fields First does
			// not populate).
			item = Item{}
			return findItem(tx, userID, itemID, &item)
		}

		var effort Effort
		switch err := tx.Where("name = ?", normalised).Where("user_id = ?", userID).First(&effort).Error; {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return fmt.Errorf("%w: %s", ErrEffortNotFound, normalised)
		case err != nil:
			return fmt.Errorf("failed to query the effort: %w", err)
		}

		if err := tx.Model(&Item{}).Where("id = ?", item.ID).UpdateColumn("effort_id", effort.ID).Error; err != nil {
			return fmt.Errorf("failed to set the effort: %w", err)
		}
		item = Item{}
		return findItem(tx, userID, itemID, &item)
	})
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// findEffort loads an effort by identifier, translating a missing row into
// ErrEffortNotFound. The query is scoped to the given user so cross-user access
// is reported as not found rather than leaking existence.
func findEffort(tx *gorm.DB, userID uint, id uint, effort *Effort) error {
	err := tx.Where("user_id = ?", userID).First(effort, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("%w: %d", ErrEffortNotFound, id)
	}
	if err != nil {
		return fmt.Errorf("failed to query the effort: %w", err)
	}

	return nil
}