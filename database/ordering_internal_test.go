package database

import (
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"
)

// seedItems creates the named items in order, triages each via MoveItem with
// bottom (so the first item is highest priority and the last is lowest), and
// returns their identifiers keyed by title. The display order therefore
// matches the creation order.
func seedItems(t *testing.T, db *gorm.DB, titles ...string) map[string]uint {
	t.Helper()

	ids := make(map[string]uint, len(titles))
	for _, title := range titles {
		item := Item{Title: title, UserID: testUserID}
		if err := db.Create(&item).Error; err != nil {
			t.Fatalf("failed to create the item %q: %v", title, err)
		}
		ids[title] = item.ID
	}

	// Triage in creation order by appending to the tail (bottom). The first
	// item lands at priority 0 on an empty ordering, the next at -priorityStep,
	// and so on, so the display order (DESC) matches the creation order.
	for _, title := range titles {
		anchor := MoveAnchor{Bottom: true}
		if _, err := MoveItem(db, testUserID, ids[title], anchor, MoveOptions{}); err != nil {
			t.Fatalf("failed to triage %q: %v", title, err)
		}
	}

	return ids
}

// activeTitles returns the titles of the triaged active items in manual order
// (highest priority first). Untriaged items are excluded; they are listed via
// ListItemsByView with ItemViewUntriaged.
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

func TestNextHighestPriority(t *testing.T) {
	db := setupTestDB(t)

	priority, err := NextHighestPriority(db, testUserID)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if priority != priorityStep {
		t.Errorf("expected %v on an empty database but got %v", priorityStep, priority)
	}

	seedItems(t, db, "a", "b", "c")

	priority, err = NextHighestPriority(db, testUserID)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	// seedItems triages via bottom: a=0, b=-priorityStep, c=-2*priorityStep,
	// so the highest is 0 and the next highest is priorityStep.
	if expected := priorityStep; priority != expected {
		t.Errorf("expected %v after three items but got %v", expected, priority)
	}
}

func TestNextLowestPriority(t *testing.T) {
	db := setupTestDB(t)

	priority, err := NextLowestPriority(db, testUserID)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if priority != 0 {
		t.Errorf("expected 0 on an empty database but got %v", priority)
	}

	seedItems(t, db, "a", "b", "c")

	priority, err = NextLowestPriority(db, testUserID)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	// seedItems triages via bottom: a=0, b=-priorityStep, c=-2*priorityStep,
	// so the lowest is -2*priorityStep and the next lowest is -3*priorityStep.
	if expected := -3 * priorityStep; priority != expected {
		t.Errorf("expected %v after three items but got %v", expected, priority)
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

func TestMoveItemAbsoluteTopBottom(t *testing.T) {
	tests := []struct {
		name     string
		subject  string
		top      bool
		expected []string
	}{
		{"top triages the first item", "a", true, []string{"a"}},
		{"bottom triages the first item", "b", false, []string{"b"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := setupTestDB(t)
			ids := seedItems(t, db, "a", "b", "c")

			// Reset a fresh untriaged item by clearing its priority directly.
			if err := db.Model(&Item{}).Where("id = ?", ids[test.subject]).
				Update("priority", nil).Error; err != nil {
				t.Fatalf("failed to clear the priority: %v", err)
			}

			anchor := MoveAnchor{Top: test.top, Bottom: !test.top}
			if _, err := MoveItem(db, testUserID, ids[test.subject], anchor, MoveOptions{}); err != nil {
				t.Fatalf("expected no error but got %v", err)
			}

			// The triaged item must carry a priority after the absolute move.
			var item Item
			if err := db.Select("priority").First(&item, ids[test.subject]).Error; err != nil {
				t.Fatalf("failed to load the item: %v", err)
			}
			if item.Priority == nil {
				t.Fatalf("expected the item to carry a priority after triage")
			}
		})
	}
}

func TestMoveItemTopOnEmptyOrdering(t *testing.T) {
	db := setupTestDB(t)

	// A single untriaged item with no prioritised peers can be triaged via top.
	item := Item{Title: "solo", UserID: testUserID}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("failed to create the item: %v", err)
	}

	moved, err := MoveItem(db, testUserID, item.ID, MoveAnchor{Top: true}, MoveOptions{})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if moved.Priority == nil || *moved.Priority != priorityStep {
		t.Errorf("expected priority %v on an empty ordering but got %v", priorityStep, moved.Priority)
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

func TestMoveItemRejectsUntriagedAnchor(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a", "b")

	// Create an untriaged item directly (no priority).
	untriaged := Item{Title: "untriaged", UserID: testUserID}
	if err := db.Create(&untriaged).Error; err != nil {
		t.Fatalf("failed to create the untriaged item: %v", err)
	}

	// Moving relative to an untriaged anchor is rejected: relative moves
	// require a triaged anchor.
	anchor := MoveAnchor{TargetID: untriaged.ID, Before: true}
	_, err := MoveItem(db, testUserID, ids["a"], anchor, MoveOptions{})
	if !errors.Is(err, ErrAnchorUntriaged) {
		t.Errorf("expected %v but got %v", ErrAnchorUntriaged, err)
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
	// priorityEpsilon after roughly thirty iterations, so sixty guarantees the
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
		if item.Priority == nil {
			t.Fatalf("expected %q to carry a priority", item.Title)
		}
	}
	// Adjacent gaps must remain splittable after all that churn. Under DESC
	// order items[i-1] has the higher priority, so the gap is the difference
	// the other way around compared to the old ASC ordering.
	for i := 1; i < len(items); i++ {
		if gap := *items[i-1].Priority - *items[i].Priority; gap < priorityEpsilon {
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
	// Under DESC order the first item (highest) gets N*priorityStep, the last
	// gets 1*priorityStep.
	for i, item := range items {
		if expected := priorityStep * float64(len(items)-i); *item.Priority != expected {
			t.Errorf("expected %q at %v but got %v", item.Title, expected, *item.Priority)
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
	if completed.Priority != nil {
		t.Errorf("expected the priority to be cleared but got %v", *completed.Priority)
	}
	if titles := activeTitles(t, db); !equalStrings(titles, []string{"a", "c"}) {
		t.Errorf("expected [a c] but got %v", titles)
	}

	reopened, err := SetDone(db, testUserID, ids["b"], false)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	// Reopening returns the item to the untriaged bucket rather than appending
	// it to the manual order, so it carries no priority and does not appear in
	// the default active listing.
	if reopened.Priority != nil {
		t.Fatalf("expected no priority after reopening the item but got %v", *reopened.Priority)
	}
	if titles := activeTitles(t, db); !equalStrings(titles, []string{"a", "c"}) {
		t.Errorf("expected [a c] with the reopened item untriaged but got %v", titles)
	}
	// The reopened item shows up in the untriaged view instead.
	untriaged, err := ListItemsByView(db, testUserID, ItemFilter{View: ItemViewUntriaged})
	if err != nil {
		t.Fatalf("failed to list the untriaged items: %v", err)
	}
	if len(untriaged) != 1 || untriaged[0].Title != "b" {
		t.Errorf("expected [b] untriaged but got %v", untriaged)
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
	ids := seedItems(t, db, "triaged-a", "triaged-b", "triaged-c")

	// An untriaged item is created directly so it never receives a priority.
	untriaged := Item{Title: "untriaged-d", UserID: testUserID}
	if err := db.Create(&untriaged).Error; err != nil {
		t.Fatalf("failed to create the untriaged item: %v", err)
	}

	// Give triaged-c a due date so it is also time-sensitive.
	due := time.Date(2026, time.August, 15, 0, 0, 0, 0, time.UTC)
	if err := db.Model(&Item{}).Where("id = ?", ids["triaged-c"]).
		UpdateColumn("due_date", due).Error; err != nil {
		t.Fatalf("failed to set the due date: %v", err)
	}

	// Complete triaged-b.
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
		{"untriaged", ItemViewUntriaged, []string{"untriaged-d"}},
		{"triaged", ItemViewTriaged, []string{"triaged-a", "triaged-c"}},
		{"time-sensitive", ItemViewTimeSensitive, []string{"triaged-c"}},
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

func TestNeighbourPriorityReturnsNilAtExtents(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a", "b", "c")

	// "a" is the head (highest priority), so there is no active item with a
	// higher priority to move it before.
	neighbour, err := neighbourPriority(db, testUserID, ids["a"], *priorityOf(t, db, ids["a"]), true)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if neighbour != nil {
		t.Errorf("expected nil before the head but got %v", *neighbour)
	}

	// "c" is the tail (lowest priority), so there is no active item with a
	// lower priority to move it after.
	neighbour, err = neighbourPriority(db, testUserID, ids["c"], *priorityOf(t, db, ids["c"]), false)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if neighbour != nil {
		t.Errorf("expected nil after the tail but got %v", *neighbour)
	}
}

// priorityOf loads the priority of the named item for use in neighbourPriority
// assertions.
func priorityOf(t *testing.T, db *gorm.DB, id uint) *float64 {
	t.Helper()
	var item Item
	if err := db.Select("priority").First(&item, id).Error; err != nil {
		t.Fatalf("failed to load the priority of item %d: %v", id, err)
	}
	return item.Priority
}