package database

import (
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"
)

// seedTodos creates the named todos in order, each appended to the tail of the
// manual ordering, and returns their identifiers keyed by title.
func seedTodos(t *testing.T, db *gorm.DB, titles ...string) map[string]uint {
	t.Helper()

	ids := make(map[string]uint, len(titles))
	for _, title := range titles {
		todo := Todo{Title: title}
		if err := AssignInitialPosition(db, &todo, testUserID); err != nil {
			t.Fatalf("failed to assign an initial position to %q: %v", title, err)
		}
		if err := db.Create(&todo).Error; err != nil {
			t.Fatalf("failed to create the todo %q: %v", title, err)
		}
		ids[title] = todo.ID
	}

	return ids
}

// activeTitles returns the titles of the active todos in manual order.
func activeTitles(t *testing.T, db *gorm.DB) []string {
	t.Helper()

	todos, err := ListActive(db, testUserID, TodoFilter{})
	if err != nil {
		t.Fatalf("failed to list the active todos: %v", err)
	}

	titles := make([]string, 0, len(todos))
	for _, todo := range todos {
		titles = append(titles, todo.Title)
	}

	return titles
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func TestNextPosition(t *testing.T) {
	db := setupTestDB(t)

	position, err := NextPosition(db, testUserID)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if position != positionStep {
		t.Errorf("expected %v on an empty database but got %v", positionStep, position)
	}

	seedTodos(t, db, "a", "b", "c")

	position, err = NextPosition(db, testUserID)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if expected := 4 * positionStep; position != expected {
		t.Errorf("expected %v after three todos but got %v", expected, position)
	}
}

func TestAssignInitialPositionAppendsToTail(t *testing.T) {
	db := setupTestDB(t)
	seedTodos(t, db, "a", "b", "c")

	if titles := activeTitles(t, db); !equalStrings(titles, []string{"a", "b", "c"}) {
		t.Errorf("expected [a b c] but got %v", titles)
	}
}

func TestAssignInitialPositionSkipsCompletedTodos(t *testing.T) {
	db := setupTestDB(t)

	todo := Todo{Title: "done already", Done: true}
	if err := AssignInitialPosition(db, &todo, testUserID); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if todo.Position != nil {
		t.Errorf("expected no position on a completed todo but got %v", *todo.Position)
	}
	if todo.UserID != testUserID {
		t.Errorf("expected the user id to be set but got %d", todo.UserID)
	}
}

func TestMoveTodo(t *testing.T) {
	tests := []struct {
		name     string
		subject  string
		anchor   string
		before   bool
		expected []string
	}{
		{"to the head", "d", "a", true, []string{"d", "a", "b", "c"}},
		{"to the tail", "a", "d", false, []string{"b", "c", "d", "a"}},
		{"forward one slot", "a", "b", false, []string{"b", "a", "c", "d"}},
		{"backward one slot", "d", "c", true, []string{"a", "b", "d", "c"}},
		{"before a distant todo", "a", "d", true, []string{"b", "c", "a", "d"}},
		{"after a distant todo", "d", "a", false, []string{"a", "d", "b", "c"}},
		{"to the slot it already occupies", "a", "b", true, []string{"a", "b", "c", "d"}},
		{"relative to itself", "b", "b", true, []string{"a", "b", "c", "d"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := setupTestDB(t)
			ids := seedTodos(t, db, "a", "b", "c", "d")

			anchor := MoveAnchor{TargetID: ids[test.anchor], Before: test.before}
			if _, err := MoveTodo(db, testUserID, ids[test.subject], anchor, MoveOptions{}); err != nil {
				t.Fatalf("expected no error but got %v", err)
			}

			if titles := activeTitles(t, db); !equalStrings(titles, test.expected) {
				t.Errorf("expected %v but got %v", test.expected, titles)
			}
		})
	}
}

func TestMoveTodoErrors(t *testing.T) {
	tests := []struct {
		name     string
		subject  string
		anchor   string
		doneList []string
		expected error
	}{
		{"missing subject", "missing", "a", nil, ErrTodoNotFound},
		{"missing anchor", "a", "missing", nil, ErrTodoNotFound},
		{"completed subject", "a", "b", []string{"a"}, ErrTodoCompleted},
		{"completed anchor", "a", "b", []string{"b"}, ErrAnchorCompleted},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := setupTestDB(t)
			ids := seedTodos(t, db, "a", "b", "c")
			// Identifiers are sequential, so an unseeded key resolves to zero
			// which never matches an existing row.
			for _, title := range test.doneList {
				if _, err := SetDone(db, testUserID, ids[title], true); err != nil {
					t.Fatalf("failed to complete %q: %v", title, err)
				}
			}

			anchor := MoveAnchor{TargetID: ids[test.anchor], Before: true}
			_, err := MoveTodo(db, testUserID, ids[test.subject], anchor, MoveOptions{})
			if !errors.Is(err, test.expected) {
				t.Errorf("expected %v but got %v", test.expected, err)
			}
		})
	}
}

func TestMoveTodoChangesList(t *testing.T) {
	db := setupTestDB(t)
	ids := seedTodos(t, db, "a", "b", "c")

	list := List{Name: "work", UserID: testUserID}
	if err := db.Create(&list).Error; err != nil {
		t.Fatalf("failed to create the list: %v", err)
	}

	anchor := MoveAnchor{TargetID: ids["a"], Before: true}
	moved, err := MoveTodo(db, testUserID, ids["c"], anchor, MoveOptions{ChangeList: true, ListID: &list.ID})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if moved.ListID == nil || *moved.ListID != list.ID {
		t.Fatalf("expected the todo to be assigned to list %d but got %v", list.ID, moved.ListID)
	}
	if titles := activeTitles(t, db); !equalStrings(titles, []string{"c", "a", "b"}) {
		t.Errorf("expected [c a b] but got %v", titles)
	}

	// Clearing the list is distinct from leaving it untouched.
	anchor = MoveAnchor{TargetID: ids["b"], Before: false}
	moved, err = MoveTodo(db, testUserID, ids["c"], anchor, MoveOptions{ChangeList: true, ListID: nil})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if moved.ListID != nil {
		t.Errorf("expected the list to be cleared but got %v", *moved.ListID)
	}
	if titles := activeTitles(t, db); !equalStrings(titles, []string{"a", "b", "c"}) {
		t.Errorf("expected [a b c] but got %v", titles)
	}
}

func TestMoveTodoLeavesListUntouched(t *testing.T) {
	db := setupTestDB(t)
	ids := seedTodos(t, db, "a", "b")

	list := List{Name: "work", UserID: testUserID}
	if err := db.Create(&list).Error; err != nil {
		t.Fatalf("failed to create the list: %v", err)
	}
	if err := db.Model(&Todo{}).Where("id = ?", ids["b"]).Update("list_id", list.ID).Error; err != nil {
		t.Fatalf("failed to assign the list: %v", err)
	}

	anchor := MoveAnchor{TargetID: ids["a"], Before: true}
	moved, err := MoveTodo(db, testUserID, ids["b"], anchor, MoveOptions{})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if moved.ListID == nil || *moved.ListID != list.ID {
		t.Errorf("expected the list to be retained but got %v", moved.ListID)
	}
}

func TestMoveTodoRebalances(t *testing.T) {
	db := setupTestDB(t)
	ids := seedTodos(t, db, "a", "b", "c", "d")

	// Alternately wedging two todos into the gap just after "a" halves that gap
	// on every move. Without a rebalance the gap would fall below
	// positionEpsilon after roughly thirty iterations, so sixty guarantees the
	// lazy rebalance path is taken.
	for i := 0; i < 60; i++ {
		for _, title := range []string{"c", "d"} {
			anchor := MoveAnchor{TargetID: ids["a"], Before: false}
			if _, err := MoveTodo(db, testUserID, ids[title], anchor, MoveOptions{}); err != nil {
				t.Fatalf("expected no error on iteration %d but got %v", i, err)
			}
		}
	}

	if titles := activeTitles(t, db); !equalStrings(titles, []string{"a", "d", "c", "b"}) {
		t.Errorf("expected [a d c b] but got %v", titles)
	}

	todos, err := ListActive(db, testUserID, TodoFilter{})
	if err != nil {
		t.Fatalf("failed to list the active todos: %v", err)
	}
	for _, todo := range todos {
		if todo.Position == nil {
			t.Fatalf("expected %q to carry a position", todo.Title)
		}
	}
	// Adjacent gaps must remain splittable after all that churn.
	for i := 1; i < len(todos); i++ {
		if gap := *todos[i].Position - *todos[i-1].Position; gap < positionEpsilon {
			t.Errorf("expected a splittable gap between %q and %q but got %v", todos[i-1].Title, todos[i].Title, gap)
		}
	}
}

func TestRebalanceRestoresEvenSpacing(t *testing.T) {
	db := setupTestDB(t)
	seedTodos(t, db, "a", "b", "c")

	if err := rebalance(db, testUserID); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	todos, err := ListActive(db, testUserID, TodoFilter{})
	if err != nil {
		t.Fatalf("failed to list the active todos: %v", err)
	}
	for i, todo := range todos {
		if expected := positionStep * float64(i+1); *todo.Position != expected {
			t.Errorf("expected %q at %v but got %v", todo.Title, expected, *todo.Position)
		}
	}
}

func TestSetDone(t *testing.T) {
	db := setupTestDB(t)
	ids := seedTodos(t, db, "a", "b", "c")

	completed, err := SetDone(db, testUserID, ids["b"], true)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if !completed.Done {
		t.Errorf("expected the todo to be completed")
	}
	if completed.Position != nil {
		t.Errorf("expected the position to be cleared but got %v", *completed.Position)
	}
	if titles := activeTitles(t, db); !equalStrings(titles, []string{"a", "c"}) {
		t.Errorf("expected [a c] but got %v", titles)
	}

	reopened, err := SetDone(db, testUserID, ids["b"], false)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if reopened.Position == nil {
		t.Fatalf("expected a position after reopening the todo")
	}
	// Reopening appends to the tail rather than restoring the original slot.
	if titles := activeTitles(t, db); !equalStrings(titles, []string{"a", "c", "b"}) {
		t.Errorf("expected [a c b] but got %v", titles)
	}
}

func TestSetDoneNotFound(t *testing.T) {
	db := setupTestDB(t)

	if _, err := SetDone(db, testUserID, 404, true); !errors.Is(err, ErrTodoNotFound) {
		t.Errorf("expected %v but got %v", ErrTodoNotFound, err)
	}
}

func TestSetDoneRejectsCrossUser(t *testing.T) {
	db := setupTestDB(t)
	ids := seedTodos(t, db, "a")

	// A different user must not be able to complete (or even see) the todo.
	if _, err := SetDone(db, testUserID+1, ids["a"], true); !errors.Is(err, ErrTodoNotFound) {
		t.Errorf("expected %v for a cross-user access but got %v", ErrTodoNotFound, err)
	}
}

func TestListCompleted(t *testing.T) {
	db := setupTestDB(t)
	ids := seedTodos(t, db, "a", "b", "c")

	for _, title := range []string{"a", "b", "c"} {
		if _, err := SetDone(db, testUserID, ids[title], true); err != nil {
			t.Fatalf("failed to complete %q: %v", title, err)
		}
	}

	// The completion timestamps are set explicitly: SQLite resolution is coarse
	// enough that three consecutive updates can otherwise tie.
	base := time.Date(2026, time.July, 31, 12, 0, 0, 0, time.UTC)
	for i, title := range []string{"a", "b", "c"} {
		updatedAt := base.Add(time.Duration(i) * time.Hour)
		if err := db.Model(&Todo{}).Where("id = ?", ids[title]).
			UpdateColumn("updated_at", updatedAt).Error; err != nil {
			t.Fatalf("failed to set the timestamp of %q: %v", title, err)
		}
	}

	todos, err := ListCompleted(db, testUserID, TodoFilter{})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	titles := make([]string, 0, len(todos))
	for _, todo := range todos {
		titles = append(titles, todo.Title)
	}
	if !equalStrings(titles, []string{"c", "b", "a"}) {
		t.Errorf("expected [c b a] but got %v", titles)
	}
}

func TestFindTodoCrossUserIsNotFound(t *testing.T) {
	db := setupTestDB(t)
	ids := seedTodos(t, db, "a")

	// A different user must not be able to load the todo; cross-user access is
	// reported as not found rather than leaking existence.
	var todo Todo
	err := findTodo(db, testUserID+1, ids["a"], &todo)
	if !errors.Is(err, ErrTodoNotFound) {
		t.Errorf("expected %v but got %v", ErrTodoNotFound, err)
	}
}

func TestNeighbourPositionReturnsNilAtExtents(t *testing.T) {
	db := setupTestDB(t)
	ids := seedTodos(t, db, "a", "b", "c")

	// "a" is the head, so there is no active todo with a smaller position.
	neighbour, err := neighbourPosition(db, testUserID, ids["a"], *positionOf(t, db, ids["a"]), true)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if neighbour != nil {
		t.Errorf("expected nil before the head but got %v", *neighbour)
	}

	// "c" is the tail, so there is no active todo with a larger position.
	neighbour, err = neighbourPosition(db, testUserID, ids["c"], *positionOf(t, db, ids["c"]), false)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if neighbour != nil {
		t.Errorf("expected nil after the tail but got %v", *neighbour)
	}
}

// positionOf loads the position of the named todo for use in neighbourPosition
// assertions.
func positionOf(t *testing.T, db *gorm.DB, id uint) *float64 {
	t.Helper()
	var todo Todo
	if err := db.Select("position").First(&todo, id).Error; err != nil {
		t.Fatalf("failed to load the position of todo %d: %v", id, err)
	}
	return todo.Position
}
