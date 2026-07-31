package database

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

var (
	// ErrLabelNotFound is returned when a label does not exist.
	ErrLabelNotFound = errors.New("label not found")
	// ErrLabelExists is returned when a label name is already taken.
	ErrLabelExists = errors.New("label already exists")
	// ErrLabelInUse is returned when deleting a label that is still attached
	// to at least one item.
	ErrLabelInUse = errors.New("label is still attached to items")
	// ErrLabelNameEmpty is returned when a label name normalises to nothing.
	ErrLabelNameEmpty = errors.New("label name must not be empty")
)

// NormaliseLabelName reduces a label name to its canonical form. It is the
// single definition of label identity: any two names with the same normalised
// form refer to the same label.
func NormaliseLabelName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// normaliseLabelNames normalises a list of names, dropping blanks and
// duplicates while preserving the order in which names were first seen.
func normaliseLabelNames(names []string) []string {
	seen := make(map[string]struct{}, len(names))
	result := make([]string, 0, len(names))

	for _, name := range names {
		normalised := NormaliseLabelName(name)
		if normalised == "" {
			continue
		}
		if _, ok := seen[normalised]; ok {
			continue
		}
		seen[normalised] = struct{}{}
		result = append(result, normalised)
	}

	return result
}

// ListLabels returns every label owned by the user, ordered by name.
func ListLabels(db *gorm.DB, userID uint) ([]Label, error) {
	var labels []Label
	if err := db.Where("user_id = ?", userID).Order("name ASC").Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to list the labels: %w", err)
	}

	return labels, nil
}

// CreateLabel creates a label explicitly. Unlike the tagging path it does not
// fall back to returning an existing label, so that an accidental duplicate is
// reported rather than silently accepted.
func CreateLabel(db *gorm.DB, userID uint, name string) (*Label, error) {
	normalised := NormaliseLabelName(name)
	if normalised == "" {
		return nil, ErrLabelNameEmpty
	}

	var label Label
	err := db.Transaction(func(tx *gorm.DB) error {
		switch err := tx.Where("name = ?", normalised).Where("user_id = ?", userID).First(&label).Error; {
		case err == nil:
			return fmt.Errorf("%w: %s", ErrLabelExists, normalised)
		case !errors.Is(err, gorm.ErrRecordNotFound):
			return fmt.Errorf("failed to query the label: %w", err)
		}

		label = Label{Name: normalised, UserID: userID}
		if err := tx.Create(&label).Error; err != nil {
			return fmt.Errorf("failed to create the label: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &label, nil
}

// RenameLabel changes the name of an existing label. Renaming a label to the
// name it already has is a no-op rather than a conflict.
func RenameLabel(db *gorm.DB, userID uint, id uint, name string) (*Label, error) {
	normalised := NormaliseLabelName(name)
	if normalised == "" {
		return nil, ErrLabelNameEmpty
	}

	var label Label
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := findLabel(tx, userID, id, &label); err != nil {
			return err
		}
		if label.Name == normalised {
			return nil
		}

		var existing Label
		switch err := tx.Where("name = ?", normalised).Where("user_id = ?", userID).First(&existing).Error; {
		case err == nil:
			return fmt.Errorf("%w: %s", ErrLabelExists, normalised)
		case !errors.Is(err, gorm.ErrRecordNotFound):
			return fmt.Errorf("failed to query the label: %w", err)
		}

		if err := tx.Model(&label).Update("name", normalised).Error; err != nil {
			return fmt.Errorf("failed to rename the label: %w", err)
		}

		return findLabel(tx, userID, id, &label)
	})
	if err != nil {
		return nil, err
	}

	return &label, nil
}

// DeleteLabel removes a label that is no longer attached to any item. Labels in
// use are reported rather than silently detached, so that deleting a label can
// never quietly change the tagging of an item.
func DeleteLabel(db *gorm.DB, userID uint, id uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var label Label
		if err := findLabel(tx, userID, id, &label); err != nil {
			return err
		}

		var count int64
		if err := tx.
			Table("item_labels").
			Joins("JOIN items ON items.id = item_labels.item_id").
			Where("item_labels.label_id = ?", id).
			Where("items.user_id = ?", userID).
			Where("items.deleted_at IS NULL").
			Count(&count).Error; err != nil {
			return fmt.Errorf("failed to count the items using the label: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("%w: %s is attached to %d item(s)", ErrLabelInUse, label.Name, count)
		}

		// Any remaining join rows belong to soft deleted items. The join table
		// has no soft delete column of its own, so they are removed here to
		// avoid leaving rows pointing at a label that no longer exists.
		if err := tx.Exec("DELETE FROM item_labels WHERE label_id = ?", id).Error; err != nil {
			return fmt.Errorf("failed to detach the label: %w", err)
		}

		if err := tx.Delete(&label).Error; err != nil {
			return fmt.Errorf("failed to delete the label: %w", err)
		}

		return nil
	})
}

// FindOrCreateLabels resolves names to labels, creating any that do not exist
// yet. Names are normalised, and blanks and duplicates are discarded.
func FindOrCreateLabels(db *gorm.DB, userID uint, names []string) ([]Label, error) {
	normalised := normaliseLabelNames(names)
	if len(normalised) == 0 {
		return nil, nil
	}

	labels := make([]Label, 0, len(normalised))
	err := db.Transaction(func(tx *gorm.DB) error {
		labels = labels[:0]
		for _, name := range normalised {
			var label Label
			err := tx.Where("name = ?", name).Where("user_id = ?", userID).First(&label).Error
			switch {
			case err == nil:
			case errors.Is(err, gorm.ErrRecordNotFound):
				label = Label{Name: name, UserID: userID}
				if createErr := tx.Create(&label).Error; createErr != nil {
					return fmt.Errorf("failed to create the label %q: %w", name, createErr)
				}
			default:
				return fmt.Errorf("failed to query the label %q: %w", name, err)
			}
			labels = append(labels, label)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return labels, nil
}

// findLabelsByName resolves names to labels that already exist, ignoring any
// that do not. It backs removal, where creating a label purely to detach it
// again would be pointless.
func findLabelsByName(db *gorm.DB, userID uint, names []string) ([]Label, error) {
	normalised := normaliseLabelNames(names)
	if len(normalised) == 0 {
		return nil, nil
	}

	var labels []Label
	if err := db.Where("name IN ?", normalised).Where("user_id = ?", userID).Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to query the labels: %w", err)
	}

	return labels, nil
}

// UpdateItemLabels attaches and detaches labels on an item. Labels being added
// are created on the fly when they do not exist; labels being removed are
// matched against existing labels only. Adding a label an item already carries
// and removing one it does not are both no-ops.
func UpdateItemLabels(db *gorm.DB, userID uint, itemID uint, add, remove []string) (*Item, error) {
	var item Item
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := findItem(tx, userID, itemID, &item); err != nil {
			return err
		}

		// Removal happens first so that passing the same name to both flags
		// leaves the label attached rather than depending on statement order.
		if len(remove) > 0 {
			labels, err := findLabelsByName(tx, userID, remove)
			if err != nil {
				return err
			}
			if len(labels) > 0 {
				if err := tx.Model(&item).Association("Labels").Delete(labels); err != nil {
					return fmt.Errorf("failed to remove the labels: %w", err)
				}
			}
		}

		if len(add) > 0 {
			labels, err := FindOrCreateLabels(tx, userID, add)
			if err != nil {
				return err
			}
			if len(labels) > 0 {
				if err := tx.Model(&item).Association("Labels").Append(labels); err != nil {
					return fmt.Errorf("failed to add the labels: %w", err)
				}
			}
		}

		return findItem(tx, userID, itemID, &item)
	})
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// findLabel loads a label by identifier, translating a missing row into
// ErrLabelNotFound. The query is scoped to the given user so cross-user access
// is reported as not found rather than leaking existence.
func findLabel(tx *gorm.DB, userID uint, id uint, label *Label) error {
	err := tx.Where("user_id = ?", userID).First(label, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("%w: %d", ErrLabelNotFound, id)
	}
	if err != nil {
		return fmt.Errorf("failed to query the label: %w", err)
	}

	return nil
}