package database

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

const (
	// positionStep is the gap between adjacent ranks after a rebalance and the
	// increment used when appending an item to the tail of the manual order.
	positionStep = 1024.0
	// positionEpsilon is the smallest gap between two neighbours that can still
	// be split by a midpoint. Below this, float64 mantissa precision is close
	// enough to exhaustion that the ordering is rebalanced instead.
	positionEpsilon = 1e-6
)

var (
	// ErrItemNotFound is returned when the item being moved does not exist.
	ErrItemNotFound = errors.New("item not found")
	// ErrItemCompleted is returned when the item being moved is completed and
	// therefore carries no position.
	ErrItemCompleted = errors.New("cannot reposition a completed item")
	// ErrAnchorCompleted is returned when the item used as the move anchor is
	// completed and therefore carries no position.
	ErrAnchorCompleted = errors.New("cannot position relative to a completed item")
)

// MoveAnchor identifies where an item should be placed relative to another item.
type MoveAnchor struct {
	TargetID uint
	// Before places the subject immediately before TargetID; otherwise it is
	// placed immediately after.
	Before bool
}

// MoveOptions carries the optional list reassignment applied in the same
// transaction as the move.
type MoveOptions struct {
	// ChangeList reports whether ListID should be applied at all. It
	// distinguishes leaving the list untouched from clearing it.
	ChangeList bool
	ListID     *uint
}

// NextPosition returns the rank to use when appending an item to the tail of the
// manual order.
func NextPosition(db *gorm.DB, userID uint) (float64, error) {
	var maxPosition *float64
	if err := db.
		Model(&Item{}).
		Where("done = ?", false).
		Where("user_id = ?", userID).
		Select("MAX(position)").
		Scan(&maxPosition).Error; err != nil {
		return 0, fmt.Errorf("failed to determine the next position: %w", err)
	}

	if maxPosition == nil {
		return positionStep, nil
	}

	return *maxPosition + positionStep, nil
}

// AssignInitialPosition sets the position of a new item so that it lands at the
// tail of the manual order. Completed items are left without a position. The
// item's UserID is set to the given identifier so callers need only assign the
// user once.
func AssignInitialPosition(db *gorm.DB, item *Item, userID uint) error {
	item.UserID = userID
	if item.Done {
		item.Position = nil
		return nil
	}

	position, err := NextPosition(db, userID)
	if err != nil {
		return err
	}
	item.Position = &position

	return nil
}

// ItemFilter narrows an item listing.
type ItemFilter struct {
	// Labels restricts the result to items carrying every one of these labels.
	// Names are normalised before matching, and an unknown name therefore
	// yields no results rather than being ignored.
	Labels []string
	// View narrows the listing to a single bucket. ItemViewUnspecified keeps
	// the legacy two-bucket behaviour served by ListActive and ListCompleted.
	View ItemView
}

// ItemView selects a bucket of items. It mirrors the proto enum so the
// database layer stays decoupled from the generated code.
type ItemView int

const (
	ItemViewUnspecified   ItemView = 0
	ItemViewUntriaged     ItemView = 1
	ItemViewTriaged       ItemView = 2
	ItemViewTimeSensitive ItemView = 3
	ItemViewDone          ItemView = 4
)

// apply narrows a query to the items matching the filter.
func (f ItemFilter) apply(db *gorm.DB) *gorm.DB {
	names := normaliseLabelNames(f.Labels)
	if len(names) == 0 {
		return db
	}

	// Every requested label has to be present, so the matching join rows are
	// counted per item and compared against the number of names requested.
	return db.
		Joins("JOIN item_labels ON item_labels.item_id = items.id").
		Joins("JOIN labels ON labels.id = item_labels.label_id AND labels.deleted_at IS NULL").
		Where("labels.name IN ?", names).
		Where("labels.user_id = items.user_id").
		Group("items.id").
		Having("COUNT(DISTINCT labels.id) = ?", len(names))
}

// ListActive returns the active items in manual order. The identifier is used
// as a tiebreak so the order stays deterministic even if two ranks ever
// collide.
func ListActive(db *gorm.DB, userID uint, filter ItemFilter) ([]Item, error) {
	var items []Item
	// Columns are qualified because the label filter joins two further tables
	// that carry columns of the same name.
	if err := filter.apply(db).
		Preload("Labels").
		Select("items.*").
		Where("items.done = ?", false).
		Where("items.user_id = ?", userID).
		Order("items.position ASC, items.id ASC").
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to list active items: %w", err)
	}

	return items, nil
}

// ListCompleted returns the completed items, most recently updated first.
// Completed items do not take part in the manual ordering.
func ListCompleted(db *gorm.DB, userID uint, filter ItemFilter) ([]Item, error) {
	var items []Item
	if err := filter.apply(db).
		Preload("Labels").
		Select("items.*").
		Where("items.done = ?", true).
		Where("items.user_id = ?", userID).
		Order("items.updated_at DESC, items.id DESC").
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to list completed items: %w", err)
	}

	return items, nil
}

// ListItemsByView returns the items in the bucket selected by filter.View.
// Only ItemViewUntriaged, ItemViewTriaged, ItemViewTimeSensitive, and
// ItemViewDone are supported here; ItemViewUnspecified is handled by the
// caller via ListActive and ListCompleted.
func ListItemsByView(db *gorm.DB, userID uint, filter ItemFilter) ([]Item, error) {
	var items []Item

	query := filter.apply(db).
		Preload("Labels").
		Select("items.*").
		Where("items.user_id = ?", userID)

	switch filter.View {
	case ItemViewUntriaged:
		query = query.
			Where("items.done = ?", false).
			Where("items.position IS NULL").
			Order("items.id ASC")
	case ItemViewTriaged:
		query = query.
			Where("items.done = ?", false).
			Where("items.position IS NOT NULL").
			Order("items.position ASC, items.id ASC")
	case ItemViewTimeSensitive:
		query = query.
			Where("items.done = ?", false).
			Where("items.due_date IS NOT NULL").
			Order("items.position ASC, items.id ASC")
	case ItemViewDone:
		query = query.
			Where("items.done = ?", true).
			Order("items.updated_at DESC, items.id DESC")
	default:
		return nil, fmt.Errorf("unsupported item view: %d", filter.View)
	}

	if err := query.Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to list items by view: %w", err)
	}

	return items, nil
}

// activeItemsForRebalance returns the active items in manual order without
// loading their labels. Rebalancing runs inside a move transaction, so it is
// kept as cheap as possible.
func activeItemsForRebalance(db *gorm.DB, userID uint) ([]Item, error) {
	var items []Item
	if err := db.
		Select("id", "position").
		Where("done = ?", false).
		Where("user_id = ?", userID).
		Order("position ASC, id ASC").
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to list active items: %w", err)
	}

	return items, nil
}

// SetDone marks an item as completed or active. Completing an item removes it
// from the manual ordering; making it active again appends it to the tail. This
// is the only place where the "an item has a position exactly when it is active"
// invariant is maintained.
func SetDone(db *gorm.DB, userID uint, id uint, done bool) (*Item, error) {
	var item Item
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := findItem(tx, userID, id, &item); err != nil {
			return err
		}

		updates := map[string]any{"done": done, "position": nil}
		if !done {
			position, err := NextPosition(tx, userID)
			if err != nil {
				return err
			}
			updates["position"] = position
		}

		if err := tx.Model(&item).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update the item: %w", err)
		}

		return findItem(tx, userID, id, &item)
	})
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// MoveItem places an item immediately before or after another item in the manual
// order and optionally reassigns its list in the same transaction.
func MoveItem(db *gorm.DB, userID uint, id uint, anchor MoveAnchor, opts MoveOptions) (*Item, error) {
	var subject Item
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := findItem(tx, userID, id, &subject); err != nil {
			return err
		}
		if subject.Done {
			return ErrItemCompleted
		}

		// Moving an item relative to itself only ever applies the list change.
		if id != anchor.TargetID {
			position, err := resolvePosition(tx, userID, id, anchor)
			if err != nil {
				return err
			}
			if err := tx.Model(&subject).Update("position", position).Error; err != nil {
				return fmt.Errorf("failed to update the position of the item: %w", err)
			}
		}

		if opts.ChangeList {
			// A map is used rather than a struct so that a nil list identifier
			// is written as NULL instead of being skipped as a zero value.
			if err := tx.Model(&subject).
				Updates(map[string]any{"list_id": opts.ListID}).Error; err != nil {
				return fmt.Errorf("failed to update the list of the item: %w", err)
			}
		}

		return findItem(tx, userID, id, &subject)
	})
	if err != nil {
		return nil, err
	}

	return &subject, nil
}

// resolvePosition computes the rank placing subjectID at the anchor,
// rebalancing the ordering once if the neighbouring gap is too small to split.
func resolvePosition(tx *gorm.DB, userID uint, subjectID uint, anchor MoveAnchor) (float64, error) {
	for range 2 {
		var target Item
		if err := findItem(tx, userID, anchor.TargetID, &target); err != nil {
			return 0, err
		}
		if target.Done || target.Position == nil {
			return 0, ErrAnchorCompleted
		}

		neighbour, err := neighbourPosition(tx, userID, subjectID, *target.Position, anchor.Before)
		if err != nil {
			return 0, err
		}

		// No neighbour on that side means the subject becomes the new head or
		// tail, so a full step past the target is always available.
		if neighbour == nil {
			if anchor.Before {
				return *target.Position - positionStep, nil
			}
			return *target.Position + positionStep, nil
		}

		// The check is on the gap the split would leave behind, not the gap
		// being split: halving a gap of 1.5 * positionEpsilon would otherwise
		// produce neighbours that can no longer be separated.
		if gap := absDiff(*target.Position, *neighbour); gap/2 >= positionEpsilon {
			return (*target.Position + *neighbour) / 2, nil
		}

		if err := rebalance(tx, userID); err != nil {
			return 0, err
		}
	}

	// Unreachable in practice: every gap is positionStep after a rebalance.
	return 0, fmt.Errorf("failed to find a position for item %d after rebalancing", subjectID)
}

// neighbourPosition returns the rank of the active item sitting immediately on
// the requested side of position, ignoring the item being moved. It returns nil
// when there is no such item.
func neighbourPosition(tx *gorm.DB, userID uint, subjectID uint, position float64, before bool) (*float64, error) {
	comparison, order := "position > ?", "position ASC"
	if before {
		comparison, order = "position < ?", "position DESC"
	}

	var neighbour Item
	err := tx.
		Where("done = ?", false).
		Where("user_id = ?", userID).
		Where("id <> ?", subjectID).
		Where(comparison, position).
		Order(order).
		First(&neighbour).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find the neighbouring item: %w", err)
	}

	return neighbour.Position, nil
}

// rebalance rewrites every active rank as an evenly spaced multiple of
// positionStep, preserving the current order and restoring the gaps that make
// midpoint insertion possible.
func rebalance(tx *gorm.DB, userID uint) error {
	items, err := activeItemsForRebalance(tx, userID)
	if err != nil {
		return err
	}

	for i := range items {
		position := positionStep * float64(i+1)
		if err := tx.Model(&Item{}).
			Where("id = ?", items[i].ID).
			Update("position", position).Error; err != nil {
			return fmt.Errorf("failed to rebalance the item ordering: %w", err)
		}
	}

	return nil
}

// findItem loads an item by identifier, translating a missing row into
// ErrItemNotFound. The query is scoped to the given user so cross-user access
// is reported as not found rather than leaking existence.
func findItem(tx *gorm.DB, userID uint, id uint, item *Item) error {
	err := tx.Preload("Labels").Where("user_id = ?", userID).First(item, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("%w: %d", ErrItemNotFound, id)
	}
	if err != nil {
		return fmt.Errorf("failed to query the item: %w", err)
	}

	return nil
}

func absDiff(a, b float64) float64 {
	if a > b {
		return a - b
	}
	return b - a
}
