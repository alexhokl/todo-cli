package database

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

const (
	// positionStep is the gap between adjacent ranks after a rebalance and the
	// increment used when appending a todo to the tail of the manual order.
	positionStep = 1024.0
	// positionEpsilon is the smallest gap between two neighbours that can still
	// be split by a midpoint. Below this, float64 mantissa precision is close
	// enough to exhaustion that the ordering is rebalanced instead.
	positionEpsilon = 1e-6
)

var (
	// ErrTodoNotFound is returned when the todo being moved does not exist.
	ErrTodoNotFound = errors.New("todo not found")
	// ErrTodoCompleted is returned when the todo being moved is completed and
	// therefore carries no position.
	ErrTodoCompleted = errors.New("cannot reposition a completed todo")
	// ErrAnchorCompleted is returned when the todo used as the move anchor is
	// completed and therefore carries no position.
	ErrAnchorCompleted = errors.New("cannot position relative to a completed todo")
)

// MoveAnchor identifies where a todo should be placed relative to another todo.
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

// NextPosition returns the rank to use when appending a todo to the tail of the
// manual order.
func NextPosition(db *gorm.DB, userID uint) (float64, error) {
	var maxPosition *float64
	if err := db.
		Model(&Todo{}).
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

// AssignInitialPosition sets the position of a new todo so that it lands at the
// tail of the manual order. Completed todos are left without a position. The
// todo's UserID is set to the given identifier so callers need only assign the
// user once.
func AssignInitialPosition(db *gorm.DB, todo *Todo, userID uint) error {
	todo.UserID = userID
	if todo.Done {
		todo.Position = nil
		return nil
	}

	position, err := NextPosition(db, userID)
	if err != nil {
		return err
	}
	todo.Position = &position

	return nil
}

// TodoFilter narrows a todo listing.
type TodoFilter struct {
	// Labels restricts the result to todos carrying every one of these labels.
	// Names are normalised before matching, and an unknown name therefore
	// yields no results rather than being ignored.
	Labels []string
}

// apply narrows a query to the todos matching the filter.
func (f TodoFilter) apply(db *gorm.DB) *gorm.DB {
	names := normaliseLabelNames(f.Labels)
	if len(names) == 0 {
		return db
	}

	// Every requested label has to be present, so the matching join rows are
	// counted per todo and compared against the number of names requested.
	return db.
		Joins("JOIN todo_labels ON todo_labels.todo_id = todos.id").
		Joins("JOIN labels ON labels.id = todo_labels.label_id AND labels.deleted_at IS NULL").
		Where("labels.name IN ?", names).
		Where("labels.user_id = todos.user_id").
		Group("todos.id").
		Having("COUNT(DISTINCT labels.id) = ?", len(names))
}

// ListActive returns the active todos in manual order. The identifier is used
// as a tiebreak so the order stays deterministic even if two ranks ever
// collide.
func ListActive(db *gorm.DB, userID uint, filter TodoFilter) ([]Todo, error) {
	var todos []Todo
	// Columns are qualified because the label filter joins two further tables
	// that carry columns of the same name.
	if err := filter.apply(db).
		Preload("Labels").
		Select("todos.*").
		Where("todos.done = ?", false).
		Where("todos.user_id = ?", userID).
		Order("todos.position ASC, todos.id ASC").
		Find(&todos).Error; err != nil {
		return nil, fmt.Errorf("failed to list active todos: %w", err)
	}

	return todos, nil
}

// ListCompleted returns the completed todos, most recently updated first.
// Completed todos do not take part in the manual ordering.
func ListCompleted(db *gorm.DB, userID uint, filter TodoFilter) ([]Todo, error) {
	var todos []Todo
	if err := filter.apply(db).
		Preload("Labels").
		Select("todos.*").
		Where("todos.done = ?", true).
		Where("todos.user_id = ?", userID).
		Order("todos.updated_at DESC, todos.id DESC").
		Find(&todos).Error; err != nil {
		return nil, fmt.Errorf("failed to list completed todos: %w", err)
	}

	return todos, nil
}

// activeTodosForRebalance returns the active todos in manual order without
// loading their labels. Rebalancing runs inside a move transaction, so it is
// kept as cheap as possible.
func activeTodosForRebalance(db *gorm.DB, userID uint) ([]Todo, error) {
	var todos []Todo
	if err := db.
		Select("id", "position").
		Where("done = ?", false).
		Where("user_id = ?", userID).
		Order("position ASC, id ASC").
		Find(&todos).Error; err != nil {
		return nil, fmt.Errorf("failed to list active todos: %w", err)
	}

	return todos, nil
}

// SetDone marks a todo as completed or active. Completing a todo removes it
// from the manual ordering; making it active again appends it to the tail. This
// is the only place where the "a todo has a position exactly when it is active"
// invariant is maintained.
func SetDone(db *gorm.DB, userID uint, id uint, done bool) (*Todo, error) {
	var todo Todo
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := findTodo(tx, userID, id, &todo); err != nil {
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

		if err := tx.Model(&todo).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update the todo: %w", err)
		}

		return findTodo(tx, userID, id, &todo)
	})
	if err != nil {
		return nil, err
	}

	return &todo, nil
}

// MoveTodo places a todo immediately before or after another todo in the manual
// order and optionally reassigns its list in the same transaction.
func MoveTodo(db *gorm.DB, userID uint, id uint, anchor MoveAnchor, opts MoveOptions) (*Todo, error) {
	var subject Todo
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := findTodo(tx, userID, id, &subject); err != nil {
			return err
		}
		if subject.Done {
			return ErrTodoCompleted
		}

		// Moving a todo relative to itself only ever applies the list change.
		if id != anchor.TargetID {
			position, err := resolvePosition(tx, userID, id, anchor)
			if err != nil {
				return err
			}
			if err := tx.Model(&subject).Update("position", position).Error; err != nil {
				return fmt.Errorf("failed to update the position of the todo: %w", err)
			}
		}

		if opts.ChangeList {
			// A map is used rather than a struct so that a nil list identifier
			// is written as NULL instead of being skipped as a zero value.
			if err := tx.Model(&subject).
				Updates(map[string]any{"list_id": opts.ListID}).Error; err != nil {
				return fmt.Errorf("failed to update the list of the todo: %w", err)
			}
		}

		return findTodo(tx, userID, id, &subject)
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
		var target Todo
		if err := findTodo(tx, userID, anchor.TargetID, &target); err != nil {
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
	return 0, fmt.Errorf("failed to find a position for todo %d after rebalancing", subjectID)
}

// neighbourPosition returns the rank of the active todo sitting immediately on
// the requested side of position, ignoring the todo being moved. It returns nil
// when there is no such todo.
func neighbourPosition(tx *gorm.DB, userID uint, subjectID uint, position float64, before bool) (*float64, error) {
	comparison, order := "position > ?", "position ASC"
	if before {
		comparison, order = "position < ?", "position DESC"
	}

	var neighbour Todo
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
		return nil, fmt.Errorf("failed to find the neighbouring todo: %w", err)
	}

	return neighbour.Position, nil
}

// rebalance rewrites every active rank as an evenly spaced multiple of
// positionStep, preserving the current order and restoring the gaps that make
// midpoint insertion possible.
func rebalance(tx *gorm.DB, userID uint) error {
	todos, err := activeTodosForRebalance(tx, userID)
	if err != nil {
		return err
	}

	for i := range todos {
		position := positionStep * float64(i+1)
		if err := tx.Model(&Todo{}).
			Where("id = ?", todos[i].ID).
			Update("position", position).Error; err != nil {
			return fmt.Errorf("failed to rebalance the todo ordering: %w", err)
		}
	}

	return nil
}

// findTodo loads a todo by identifier, translating a missing row into
// ErrTodoNotFound. The query is scoped to the given user so cross-user access
// is reported as not found rather than leaking existence.
func findTodo(tx *gorm.DB, userID uint, id uint, todo *Todo) error {
	err := tx.Preload("Labels").Where("user_id = ?", userID).First(todo, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("%w: %d", ErrTodoNotFound, id)
	}
	if err != nil {
		return fmt.Errorf("failed to query the todo: %w", err)
	}

	return nil
}

func absDiff(a, b float64) float64 {
	if a > b {
		return a - b
	}
	return b - a
}
