package database

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

const (
	// priorityStep is the gap between adjacent ranks after a rebalance and the
	// increment used when assigning an absolute priority at the head or tail of
	// the manual order.
	priorityStep = 1024.0
	// priorityEpsilon is the smallest gap between two neighbours that can still
	// be split by a midpoint. Below this, float64 mantissa precision is close
	// enough to exhaustion that the ordering is rebalanced instead.
	priorityEpsilon = 1e-6
)

var (
	// ErrItemNotFound is returned when the item being moved does not exist.
	ErrItemNotFound = errors.New("item not found")
	// ErrItemCompleted is returned when the item being moved is completed and
	// therefore carries no priority.
	ErrItemCompleted = errors.New("cannot reprioritise a completed item")
	// ErrAnchorCompleted is returned when the item used as the move anchor is
	// completed and therefore carries no priority.
	ErrAnchorCompleted = errors.New("cannot prioritise relative to a completed item")
	// ErrAnchorUntriaged is returned when the item used as the move anchor has
	// not been triaged yet and therefore carries no priority. Relative moves
	// require a triaged anchor; untriaged items must be triaged via top/bottom.
	ErrAnchorUntriaged = errors.New("cannot move relative to an untriaged item")
)

// MoveAnchor identifies where an item should be placed relative to another
// item, or at an absolute end of the ordering.
type MoveAnchor struct {
	// Relative mode: TargetID is the anchor item and Before controls the side.
	TargetID uint
	// Before places the subject immediately before TargetID (towards the head,
	// i.e. a higher priority); otherwise it is placed immediately after
	// (towards the tail, i.e. a lower priority).
	Before bool

	// Absolute mode: when Top is true the subject is assigned the highest
	// priority; when Bottom is true it is assigned the lowest. Exactly one of
	// Top/Bottom is set in absolute mode, and TargetID is zero. Absolute mode is
	// the only way to triage an item when no prioritised anchor exists yet.
	Top    bool
	Bottom bool
}

// IsAbsolute reports whether the anchor selects an absolute end rather than a
// position relative to another item.
func (a MoveAnchor) IsAbsolute() bool { return a.Top || a.Bottom }

// MoveOptions carries the optional list reassignment applied in the same
// transaction as the move.
type MoveOptions struct {
	// ChangeList reports whether ListID should be applied at all. It
	// distinguishes leaving the list untouched from clearing it.
	ChangeList bool
	ListID     *uint
}

// NextHighestPriority returns the priority to use when placing an item at the
// head of the manual order (the highest priority). It is priorityStep above the
// current maximum, or priorityStep on an empty ordering.
func NextHighestPriority(db *gorm.DB, userID uint) (float64, error) {
	var maxPriority *float64
	if err := db.
		Model(&Item{}).
		Where("done = ?", false).
		Where("user_id = ?", userID).
		Where("priority IS NOT NULL").
		Select("MAX(priority)").
		Scan(&maxPriority).Error; err != nil {
		return 0, fmt.Errorf("failed to determine the next highest priority: %w", err)
	}

	if maxPriority == nil {
		return priorityStep, nil
	}

	return *maxPriority + priorityStep, nil
}

// NextLowestPriority returns the priority to use when placing an item at the
// tail of the manual order (the lowest priority). It is priorityStep below the
// current minimum, or 0 on an empty ordering (so the first item is neutral and
// subsequent appends go negative).
func NextLowestPriority(db *gorm.DB, userID uint) (float64, error) {
	var minPriority *float64
	if err := db.
		Model(&Item{}).
		Where("done = ?", false).
		Where("user_id = ?", userID).
		Where("priority IS NOT NULL").
		Select("MIN(priority)").
		Scan(&minPriority).Error; err != nil {
		return 0, fmt.Errorf("failed to determine the next lowest priority: %w", err)
	}

	if minPriority == nil {
		return 0, nil
	}

	return *minPriority - priorityStep, nil
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

// ListActive returns the triaged active items in manual order, highest
// priority first. Untriaged items (priority IS NULL) are excluded; they are
// listed via ListItemsByView with ItemViewUntriaged. The identifier is used as a
// tiebreak so the order stays deterministic even if two ranks ever collide.
func ListActive(db *gorm.DB, userID uint, filter ItemFilter) ([]Item, error) {
	var items []Item
	// Columns are qualified because the label filter joins two further tables
	// that carry columns of the same name.
	if err := filter.apply(db).
		Preload("Labels").
		Preload("Effort").
		Preload("Blockers").
		Preload("LinkedItems", "deleted_at IS NULL").
		Select("items.*").
		Where("items.done = ?", false).
		Where("items.user_id = ?", userID).
		Where("items.priority IS NOT NULL").
		Order("items.priority DESC, items.id ASC").
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
		Preload("Effort").
		Preload("Blockers").
		Preload("LinkedItems", "deleted_at IS NULL").
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
		Preload("Effort").
		Preload("Blockers").
		Preload("LinkedItems", "deleted_at IS NULL").
		Select("items.*").
		Where("items.user_id = ?", userID)

	switch filter.View {
	case ItemViewUntriaged:
		query = query.
			Where("items.done = ?", false).
			Where("items.priority IS NULL").
			Order("items.id ASC")
	case ItemViewTriaged:
		query = query.
			Where("items.done = ?", false).
			Where("items.priority IS NOT NULL").
			Order("items.priority DESC, items.id ASC")
	case ItemViewTimeSensitive:
		query = query.
			Where("items.done = ?", false).
			Where("items.due_date IS NOT NULL").
			Order("items.priority DESC, items.id ASC")
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

// activeItemsForRebalance returns the triaged active items in manual order
// without loading their labels. Rebalancing runs inside a move transaction, so
// it is kept as cheap as possible.
func activeItemsForRebalance(db *gorm.DB, userID uint) ([]Item, error) {
	var items []Item
	if err := db.
		Select("id", "priority").
		Where("done = ?", false).
		Where("user_id = ?", userID).
		Where("priority IS NOT NULL").
		Order("priority DESC, id ASC").
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to list active items: %w", err)
	}

	return items, nil
}

// SetDone marks an item as completed or active. Completing an item removes it
// from the manual ordering; making it active again returns it to the untriaged
// bucket (priority is nil) so it can be re-prioritised rather than silently
// reappearing at the tail. This is the only place where the "an item has a
// priority exactly when it is active and triaged" invariant is maintained.
func SetDone(db *gorm.DB, userID uint, id uint, done bool) (*Item, error) {
	var item Item
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := findItem(tx, userID, id, &item); err != nil {
			return err
		}

		// Completing clears the priority; reopening leaves it nil so the item
		// becomes untriaged and must be re-prioritised explicitly.
		updates := map[string]any{"done": done, "priority": nil}

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

// MoveItem places an item at an absolute end of the manual order (top/bottom)
// or immediately before/after another item, and optionally reassigns its list
// in the same transaction. Absolute mode triages an untriaged item; relative
// mode requires the anchor to already carry a priority.
func MoveItem(db *gorm.DB, userID uint, id uint, anchor MoveAnchor, opts MoveOptions) (*Item, error) {
	var subject Item
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := findItem(tx, userID, id, &subject); err != nil {
			return err
		}
		if subject.Done {
			return ErrItemCompleted
		}

		if anchor.IsAbsolute() {
			priority, err := resolveAbsolutePriority(tx, userID, anchor)
			if err != nil {
				return err
			}
			if err := tx.Model(&subject).Update("priority", priority).Error; err != nil {
				return fmt.Errorf("failed to update the priority of the item: %w", err)
			}
		} else if id != anchor.TargetID {
			// Moving an item relative to itself only ever applies the list
			// change, so the priority rewrite is skipped in that case.
			priority, err := resolvePriority(tx, userID, id, anchor)
			if err != nil {
				return err
			}
			if err := tx.Model(&subject).Update("priority", priority).Error; err != nil {
				return fmt.Errorf("failed to update the priority of the item: %w", err)
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

// resolveAbsolutePriority computes the priority for a top/bottom placement. It
// is the only way to triage an item when no prioritised anchor exists yet.
func resolveAbsolutePriority(tx *gorm.DB, userID uint, anchor MoveAnchor) (float64, error) {
	if anchor.Top {
		return NextHighestPriority(tx, userID)
	}
	return NextLowestPriority(tx, userID)
}

// resolvePriority computes the rank placing subjectID at the relative anchor,
// rebalancing the ordering once if the neighbouring gap is too small to split.
// The anchor must already carry a priority.
func resolvePriority(tx *gorm.DB, userID uint, subjectID uint, anchor MoveAnchor) (float64, error) {
	for range 2 {
		var target Item
		if err := findItem(tx, userID, anchor.TargetID, &target); err != nil {
			return 0, err
		}
		if target.Done {
			return 0, ErrAnchorCompleted
		}
		if target.Priority == nil {
			return 0, ErrAnchorUntriaged
		}

		neighbour, err := neighbourPriority(tx, userID, subjectID, *target.Priority, anchor.Before)
		if err != nil {
			return 0, err
		}

		// No neighbour on that side means the subject becomes the new head or
		// tail, so a full step past the target is always available. "Before"
		// moves towards the head (higher priority), "after" towards the tail.
		if neighbour == nil {
			if anchor.Before {
				return *target.Priority + priorityStep, nil
			}
			return *target.Priority - priorityStep, nil
		}

		// The check is on the gap the split would leave behind, not the gap
		// being split: halving a gap of 1.5 * priorityEpsilon would otherwise
		// produce neighbours that can no longer be separated.
		if gap := absDiff(*target.Priority, *neighbour); gap/2 >= priorityEpsilon {
			return (*target.Priority + *neighbour) / 2, nil
		}

		if err := rebalance(tx, userID); err != nil {
			return 0, err
		}
	}

	// Unreachable in practice: every gap is priorityStep after a rebalance.
	return 0, fmt.Errorf("failed to find a priority for item %d after rebalancing", subjectID)
}

// neighbourPriority returns the priority of the active item sitting immediately
// on the requested side of the given priority, ignoring the item being moved.
// It returns nil when there is no such item. "Before" looks towards the head
// (higher priority), "after" towards the tail (lower priority).
func neighbourPriority(tx *gorm.DB, userID uint, subjectID uint, priority float64, before bool) (*float64, error) {
	// "Before" the anchor means towards the head, i.e. items with a higher
	// priority; the immediate predecessor is the smallest of those. "After"
	// means towards the tail, i.e. items with a lower priority; the immediate
	// successor is the largest of those.
	comparison, order := "priority < ?", "priority DESC"
	if before {
		comparison, order = "priority > ?", "priority ASC"
	}

	var neighbour Item
	err := tx.
		Where("done = ?", false).
		Where("user_id = ?", userID).
		Where("id <> ?", subjectID).
		Where(comparison, priority).
		Order(order).
		First(&neighbour).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find the neighbouring item: %w", err)
	}

	return neighbour.Priority, nil
}

// rebalance rewrites every active rank as an evenly spaced multiple of
// priorityStep, preserving the current order and restoring the gaps that make
// midpoint insertion possible. Under DESC order the first item (highest
// priority) gets the largest value.
func rebalance(tx *gorm.DB, userID uint) error {
	items, err := activeItemsForRebalance(tx, userID)
	if err != nil {
		return err
	}

	for i := range items {
		// items[0] is the highest priority; it gets N*step, the last gets step.
		priority := priorityStep * float64(len(items)-i)
		if err := tx.Model(&Item{}).
			Where("id = ?", items[i].ID).
			Update("priority", priority).Error; err != nil {
			return fmt.Errorf("failed to rebalance the item ordering: %w", err)
		}
	}

	return nil
}

// findItem loads an item by identifier, translating a missing row into
// ErrItemNotFound. The query is scoped to the given user so cross-user access
// is reported as not found rather than leaking existence.
func findItem(tx *gorm.DB, userID uint, id uint, item *Item) error {
	err := tx.Preload("Labels").Preload("Effort").Preload("Blockers").Preload("LinkedItems", "deleted_at IS NULL").Where("user_id = ?", userID).First(item, id).Error
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