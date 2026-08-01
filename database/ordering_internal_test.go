package database

import (
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"
)

// seedItems creates the named items in order, each appended to the tail of the
// manual ordering, and returns their identifiers keyed by title.
func seedItems(t *testing.T, db *gorm.DB, titles ...string) map[string]uint {
	t.Helper()

	ids := make(map[string]uint, len(titles))
	for _, title := range titles {
		item := Item{Title: title}
		if err := AssignInitialPosition(db, &item, testUserID); err != nil {
			t.Fatalf("failed to assign an initial position to %q: %v", title, err)
		}
		if err := db.Create(&item).Error; err != nil {
			t.Fatalf("failed to create the item %q: %v", title, err)
		}
		ids[title] = item.ID
	}

	return ids
}

// activeTitles returns the titles of the active items in manual order.
func activeTitles(t *testing.T, db *gorm.DB) []string {
	t.Helper()

	items, err := ListActive(db, testUserID, ItemFilter{})
	if err != nil {
		t.Fatalf("failed to list the active items: %v", err)
	}

	titles := make([]string, 0, len(items))
	for _, item := range items {
		titles = append(titles, item.Title)
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

	seedItems(t, db, "a", "b", "c")

	position, err = NextPosition(db, testUserID)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if expected := 4 * positionStep; position != expected {
		t.Errorf("expected %v after three items but got %v", expected, position)
	}
}

func TestAssignInitialPositionAppendsToTail(t *testing.T) {
	db := setupTestDB(t)
	seedItems(t, db, "a", "b", "c")

	if titles := activeTitles(t, db); !equalStrings(titles, []string{"a", "b", "c"}) {
		t.Errorf("expected [a b c] but got %v", titles)
	}
}

func TestAssignInitialPositionSkipsCompletedItems(t *testing.T) {
	db := setupTestDB(t)

	item := Item{Title: "done already", Done: true}
	if err := AssignInitialPosition(db, &item, testUserID); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if item.Position != nil {
		t.Errorf("expected no position on a completed item but got %v", *item.Position)
	}
	if item.UserID != testUserID {
		t.Errorf("expected the user id to be set but got %d", item.UserID)
	}
}

func TestMoveItem(t *testing.T) {
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
		{"before a distant item", "a", "d", true, []string{"b", "c", "a", "d"}},
		{"after a distant item", "d", "a", false, []string{"a", "d", "b", "c"}},
		{"to the slot it already occupies", "a", "b", true, []string{"a", "b", "c", "d"}},
		{"relative to itself", "b", "b", true, []string{"a", "b", "c", "d"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := setupTestDB(t)
			ids := seedItems(t, db, "a", "b", "c", "d")

			anchor := MoveAnchor{TargetID: ids[test.anchor], Before: test.before}
			if _, err := MoveItem(db, testUserID, ids[test.subject], anchor, MoveOptions{}); err != nil {
				t.Fatalf("expected no error but got %v", err)
			}

			if titles := activeTitles(t, db); !equalStrings(titles, test.expected) {
				t.Errorf("expected %v but got %v", test.expected, titles)
			}
		})
	}
}

func TestMoveItemErrors(t *testing.T) {
	tests := []struct {
		name     string
		subject  string
		anchor   string
		doneList []string
		expected error
	}{
		{"missing subject", "missing", "a", nil, ErrItemNotFound},
		{"missing anchor", "a", "missing", nil, ErrItemNotFound},
		{"completed subject", "a", "b", []string{"a"}, ErrItemCompleted},
		{"completed anchor", "a", "b", []string{"b"}, ErrAnchorCompleted},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := setupTestDB(t)
			ids := seedItems(t, db, "a", "b", "c")
			// Identifiers are sequential, so an unseeded key resolves to zero
			// which never matches an existing row.
			for _, title := range test.doneList {
				if _, err := SetDone(db, testUserID, ids[title], true); err != nil {
					t.Fatalf("failed to complete %q: %v", title, err)
				}
			}

			anchor := MoveAnchor{TargetID: ids[test.anchor], Before: true}
			_, err := MoveItem(db, testUserID, ids[test.subject], anchor, MoveOptions{})
			if !errors.Is(err, test.expected) {
				t.Errorf("expected %v but got %v", test.expected, err)
			}
		})
	}
}

func TestMoveItemChangesList(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a", "b", "c")

	list := List{Name: "work", UserID: testUserID}
	if err := db.Create(&list).Error; err != nil {
		t.Fatalf("failed to create the list: %v", err)
	}

	anchor := MoveAnchor{TargetID: ids["a"], Before: true}
	moved, err := MoveItem(db, testUserID, ids["c"], anchor, MoveOptions{ChangeList: true, ListID: &list.ID})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if moved.ListID == nil || *moved.ListID != list.ID {
		t.Fatalf("expected the item to be assigned to list %d but got %v", list.ID, moved.ListID)
	}
	if titles := activeTitles(t, db); !equalStrings(titles, []string{"c", "a", "b"}) {
		t.Errorf("expected [c a b] but got %v", titles)
	}

	// Clearing the list is distinct from leaving it untouched.
	anchor = MoveAnchor{TargetID: ids["b"], Before: false}
	moved, err = MoveItem(db, testUserID, ids["c"], anchor, MoveOptions{ChangeList: true, ListID: nil})
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

func TestMoveItemLeavesListUntouched(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a", "b")

	list := List{Name: "work", UserID: testUserID}
	if err := db.Create(&list).Error; err != nil {
		t.Fatalf("failed to create the list: %v", err)
	}
	if err := db.Model(&Item{}).Where("id = ?", ids["b"]).Update("list_id", list.ID).Error; err != nil {
		t.Fatalf("failed to assign the list: %v", err)
	}

	anchor := MoveAnchor{TargetID: ids["a"], Before: true}
	moved, err := MoveItem(db, testUserID, ids["b"], anchor, MoveOptions{})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if moved.ListID == nil || *moved.ListID != list.ID {
		t.Errorf("expected the list to be retained but got %v", moved.ListID)
	}
}

func TestMoveItemRebalances(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a", "b", "c", "d")

	// Alternately wedging two items into the gap just after "a" halves that gap
	// on every move. Without a rebalance the gap would fall below
	// positionEpsilon after roughly thirty iterations, so sixty guarantees the
	// lazy rebalance path is taken.
	for i := 0; i < 60; i++ {
		for _, title := range []string{"c", "d"} {
			anchor := MoveAnchor{TargetID: ids["a"], Before: false}
			if _, err := MoveItem(db, testUserID, ids[title], anchor, MoveOptions{}); err != nil {
				t.Fatalf("expected no error on iteration %d but got %v", i, err)
			}
		}
	}

	if titles := activeTitles(t, db); !equalStrings(titles, []string{"a", "d", "c", "b"}) {
		t.Errorf("expected [a d c b] but got %v", titles)
	}

	items, err := ListActive(db, testUserID, ItemFilter{})
	if err != nil {
		t.Fatalf("failed to list the active items: %v", err)
	}
	for _, item := range items {
		if item.Position == nil {
			t.Fatalf("expected %q to carry a position", item.Title)
		}
	}
	// Adjacent gaps must remain splittable after all that churn.
	for i := 1; i < len(items); i++ {
		if gap := *items[i].Position - *items[i-1].Position; gap < positionEpsilon {
			t.Errorf("expected a splittable gap between %q and %q but got %v", items[i-1].Title, items[i].Title, gap)
		}
	}
}

func TestRebalanceRestoresEvenSpacing(t *testing.T) {
	db := setupTestDB(t)
	seedItems(t, db, "a", "b", "c")

	if err := rebalance(db, testUserID); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	items, err := ListActive(db, testUserID, ItemFilter{})
	if err != nil {
		t.Fatalf("failed to list the active items: %v", err)
	}
	for i, item := range items {
		if expected := positionStep * float64(i+1); *item.Position != expected {
			t.Errorf("expected %q at %v but got %v", item.Title, expected, *item.Position)
		}
	}
}

func TestSetDone(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a", "b", "c")

	completed, err := SetDone(db, testUserID, ids["b"], true)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if !completed.Done {
		t.Errorf("expected the item to be completed")
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
		t.Fatalf("expected a position after reopening the item")
	}
	// Reopening appends to the tail rather than restoring the original slot.
	if titles := activeTitles(t, db); !equalStrings(titles, []string{"a", "c", "b"}) {
		t.Errorf("expected [a c b] but got %v", titles)
	}
}

func TestSetDoneNotFound(t *testing.T) {
	db := setupTestDB(t)

	if _, err := SetDone(db, testUserID, 404, true); !errors.Is(err, ErrItemNotFound) {
		t.Errorf("expected %v but got %v", ErrItemNotFound, err)
	}
}

func TestSetDoneRejectsCrossUser(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a")

	// A different user must not be able to complete (or even see) the item.
	if _, err := SetDone(db, testUserID+1, ids["a"], true); !errors.Is(err, ErrItemNotFound) {
		t.Errorf("expected %v for a cross-user access but got %v", ErrItemNotFound, err)
	}
}

func TestListCompleted(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a", "b", "c")

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
		if err := db.Model(&Item{}).Where("id = ?", ids[title]).
			UpdateColumn("updated_at", updatedAt).Error; err != nil {
			t.Fatalf("failed to set the timestamp of %q: %v", title, err)
		}
	}

	items, err := ListCompleted(db, testUserID, ItemFilter{})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	titles := make([]string, 0, len(items))
	for _, item := range items {
		titles = append(titles, item.Title)
	}
	if !equalStrings(titles, []string{"c", "b", "a"}) {
		t.Errorf("expected [c b a] but got %v", titles)
	}
}

func TestListItemsByView(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "triaged-a", "triaged-b")

	// An untriaged item is created directly without AssignInitialPosition so it
	// carries no position.
	untriaged := Item{Title: "untriaged-c", UserID: testUserID}
	if err := db.Create(&untriaged).Error; err != nil {
		t.Fatalf("failed to create the untriaged item: %v", err)
	}

	// A time-sensitive item carries a due date.
	due := time.Date(2026, time.August, 15, 0, 0, 0, 0, time.UTC)
	timeSensitive := Item{Title: "time-sensitive-d", UserID: testUserID, DueDate: &due}
	if err := AssignInitialPosition(db, &timeSensitive, testUserID); err != nil {
		t.Fatalf("failed to assign a position: %v", err)
	}
	if err := db.Create(&timeSensitive).Error; err != nil {
		t.Fatalf("failed to create the time-sensitive item: %v", err)
	}

	// Complete one triaged item.
	if _, err := SetDone(db, testUserID, ids["triaged-b"], true); err != nil {
		t.Fatalf("failed to complete the item: %v", err)
	}
	// Pin its updated_at so the done ordering is deterministic.
	base := time.Date(2026, time.July, 31, 12, 0, 0, 0, time.UTC)
	if err := db.Model(&Item{}).Where("id = ?", ids["triaged-b"]).
		UpdateColumn("updated_at", base).Error; err != nil {
		t.Fatalf("failed to set the timestamp: %v", err)
	}

	tests := []struct {
		name     string
		view     ItemView
		expected []string
	}{
		{"untriaged", ItemViewUntriaged, []string{"untriaged-c"}},
		{"triaged", ItemViewTriaged, []string{"triaged-a", "time-sensitive-d"}},
		{"time-sensitive", ItemViewTimeSensitive, []string{"time-sensitive-d"}},
		{"done", ItemViewDone, []string{"triaged-b"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			items, err := ListItemsByView(db, testUserID, ItemFilter{View: test.view})
			if err != nil {
				t.Fatalf("expected no error but got %v", err)
			}

			titles := make([]string, 0, len(items))
			for _, item := range items {
				titles = append(titles, item.Title)
			}
			if !equalStrings(titles, test.expected) {
				t.Errorf("expected %v but got %v", test.expected, titles)
			}
		})
	}
}

func TestListItemsByViewRejectsUnspecified(t *testing.T) {
	db := setupTestDB(t)

	if _, err := ListItemsByView(db, testUserID, ItemFilter{View: ItemViewUnspecified}); err == nil {
		t.Errorf("expected an error for the unspecified view but got nil")
	}
}

func TestListItemsByViewScopedPerUser(t *testing.T) {
	db := setupTestDB(t)
	seedItems(t, db, "mine")

	// A second user's untriaged item must not appear.
	other := Item{Title: "theirs", UserID: testUserID + 1}
	if err := db.Create(&other).Error; err != nil {
		t.Fatalf("failed to create the other user's item: %v", err)
	}

	items, err := ListItemsByView(db, testUserID, ItemFilter{View: ItemViewUntriaged})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected no untriaged items for the test user but got %v", items)
	}
}

func TestFindItemCrossUserIsNotFound(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a")

	// A different user must not be able to load the item; cross-user access is
	// reported as not found rather than leaking existence.
	var item Item
	err := findItem(db, testUserID+1, ids["a"], &item)
	if !errors.Is(err, ErrItemNotFound) {
		t.Errorf("expected %v but got %v", ErrItemNotFound, err)
	}
}

func TestNeighbourPositionReturnsNilAtExtents(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a", "b", "c")

	// "a" is the head, so there is no active item with a smaller position.
	neighbour, err := neighbourPosition(db, testUserID, ids["a"], *positionOf(t, db, ids["a"]), true)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if neighbour != nil {
		t.Errorf("expected nil before the head but got %v", *neighbour)
	}

	// "c" is the tail, so there is no active item with a larger position.
	neighbour, err = neighbourPosition(db, testUserID, ids["c"], *positionOf(t, db, ids["c"]), false)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if neighbour != nil {
		t.Errorf("expected nil after the tail but got %v", *neighbour)
	}
}

// positionOf loads the position of the named item for use in neighbourPosition
// assertions.
func positionOf(t *testing.T, db *gorm.DB, id uint) *float64 {
	t.Helper()
	var item Item
	if err := db.Select("position").First(&item, id).Error; err != nil {
		t.Fatalf("failed to load the position of item %d: %v", id, err)
	}
	return item.Position
}
