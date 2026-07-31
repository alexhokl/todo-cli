package database

import (
	"errors"
	"testing"

	"gorm.io/gorm"
)

// labelNames returns the names of the labels attached to an item.
func labelNames(t *testing.T, db *gorm.DB, id uint) []string {
	t.Helper()

	var item Item
	if err := findItem(db, testUserID, id, &item); err != nil {
		t.Fatalf("failed to load the item: %v", err)
	}

	names := make([]string, 0, len(item.Labels))
	for _, label := range item.Labels {
		names = append(names, label.Name)
	}

	return names
}

func containsAll(names []string, expected ...string) bool {
	index := make(map[string]struct{}, len(names))
	for _, name := range names {
		index[name] = struct{}{}
	}
	if len(index) != len(expected) {
		return false
	}
	for _, name := range expected {
		if _, ok := index[name]; !ok {
			return false
		}
	}

	return true
}

func TestNormaliseLabelName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"work", "work"},
		{"Work", "work"},
		{"WORK", "work"},
		{"  work  ", "work"},
		{"\tWork\n", "work"},
		{"two words", "two words"},
		{"", ""},
		{"   ", ""},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			if result := NormaliseLabelName(test.input); result != test.expected {
				t.Errorf("expected %q but got %q", test.expected, result)
			}
		})
	}
}

func TestNormaliseLabelNamesDropsBlanksAndDuplicates(t *testing.T) {
	result := normaliseLabelNames([]string{"Work", "  work ", "", "   ", "home", "HOME"})
	if !equalStrings(result, []string{"work", "home"}) {
		t.Errorf("expected [work home] but got %v", result)
	}
}

func TestCreateLabel(t *testing.T) {
	db := setupTestDB(t)

	label, err := CreateLabel(db, testUserID, "  Work  ")
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if label.Name != "work" {
		t.Errorf("expected the name to be normalised to %q but got %q", "work", label.Name)
	}

	// A duplicate is reported rather than quietly returning the existing label.
	if _, err := CreateLabel(db, testUserID, "WORK"); !errors.Is(err, ErrLabelExists) {
		t.Errorf("expected %v but got %v", ErrLabelExists, err)
	}

	if _, err := CreateLabel(db, testUserID, "   "); !errors.Is(err, ErrLabelNameEmpty) {
		t.Errorf("expected %v but got %v", ErrLabelNameEmpty, err)
	}
}

func TestCreateLabelIsPerUser(t *testing.T) {
	db := setupTestDB(t)

	if _, err := CreateLabel(db, testUserID, "work"); err != nil {
		t.Fatalf("failed to create the label for the first user: %v", err)
	}
	// A second user can create a label of the same name without a conflict.
	if _, err := CreateLabel(db, testUserID+1, "work"); err != nil {
		t.Errorf("expected per-user uniqueness but got %v", err)
	}
}

func TestListLabels(t *testing.T) {
	db := setupTestDB(t)
	for _, name := range []string{"work", "admin", "home"} {
		if _, err := CreateLabel(db, testUserID, name); err != nil {
			t.Fatalf("failed to create the label %q: %v", name, err)
		}
	}

	labels, err := ListLabels(db, testUserID)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	names := make([]string, 0, len(labels))
	for _, label := range labels {
		names = append(names, label.Name)
	}
	if !equalStrings(names, []string{"admin", "home", "work"}) {
		t.Errorf("expected [admin home work] but got %v", names)
	}
}

func TestListLabelsIsPerUser(t *testing.T) {
	db := setupTestDB(t)
	if _, err := CreateLabel(db, testUserID, "work"); err != nil {
		t.Fatalf("failed to create the label: %v", err)
	}
	if _, err := CreateLabel(db, testUserID+1, "admin"); err != nil {
		t.Fatalf("failed to create the label: %v", err)
	}

	if labels, err := ListLabels(db, testUserID); err != nil {
		t.Fatalf("expected no error but got %v", err)
	} else if len(labels) != 1 || labels[0].Name != "work" {
		t.Errorf("expected only [work] for the first user but got %v", labels)
	}
	if labels, err := ListLabels(db, testUserID+1); err != nil {
		t.Fatalf("expected no error but got %v", err)
	} else if len(labels) != 1 || labels[0].Name != "admin" {
		t.Errorf("expected only [admin] for the second user but got %v", labels)
	}
}

func TestRenameLabel(t *testing.T) {
	db := setupTestDB(t)
	work, err := CreateLabel(db, testUserID, "work")
	if err != nil {
		t.Fatalf("failed to create the label: %v", err)
	}
	if _, err := CreateLabel(db, testUserID, "home"); err != nil {
		t.Fatalf("failed to create the label: %v", err)
	}

	renamed, err := RenameLabel(db, testUserID, work.ID, "  Office  ")
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if renamed.Name != "office" {
		t.Errorf("expected %q but got %q", "office", renamed.Name)
	}

	// Renaming to the name the label already carries is not a conflict.
	if _, err := RenameLabel(db, testUserID, work.ID, "OFFICE"); err != nil {
		t.Errorf("expected no error on a self rename but got %v", err)
	}

	if _, err := RenameLabel(db, testUserID, work.ID, "home"); !errors.Is(err, ErrLabelExists) {
		t.Errorf("expected %v but got %v", ErrLabelExists, err)
	}

	if _, err := RenameLabel(db, testUserID, 404, "anything"); !errors.Is(err, ErrLabelNotFound) {
		t.Errorf("expected %v but got %v", ErrLabelNotFound, err)
	}

	if _, err := RenameLabel(db, testUserID, work.ID, "  "); !errors.Is(err, ErrLabelNameEmpty) {
		t.Errorf("expected %v but got %v", ErrLabelNameEmpty, err)
	}
}

func TestDeleteLabel(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a")
	label, err := CreateLabel(db, testUserID, "work")
	if err != nil {
		t.Fatalf("failed to create the label: %v", err)
	}

	if _, err := UpdateItemLabels(db, testUserID, ids["a"], []string{"work"}, nil); err != nil {
		t.Fatalf("failed to attach the label: %v", err)
	}

	// A label in use is reported rather than silently detached.
	if err := DeleteLabel(db, testUserID, label.ID); !errors.Is(err, ErrLabelInUse) {
		t.Errorf("expected %v but got %v", ErrLabelInUse, err)
	}

	if _, err := UpdateItemLabels(db, testUserID, ids["a"], nil, []string{"work"}); err != nil {
		t.Fatalf("failed to detach the label: %v", err)
	}
	if err := DeleteLabel(db, testUserID, label.ID); err != nil {
		t.Errorf("expected no error but got %v", err)
	}

	if err := DeleteLabel(db, testUserID, 404); !errors.Is(err, ErrLabelNotFound) {
		t.Errorf("expected %v but got %v", ErrLabelNotFound, err)
	}
}

func TestDeleteLabelSweepsSoftDeletedItemJoinRows(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a")
	label, err := CreateLabel(db, testUserID, "work")
	if err != nil {
		t.Fatalf("failed to create the label: %v", err)
	}
	if _, err := UpdateItemLabels(db, testUserID, ids["a"], []string{"work"}, nil); err != nil {
		t.Fatalf("failed to attach the label: %v", err)
	}

	// The join table has no soft delete column, so the row outlives the item.
	if err := db.Delete(&Item{}, ids["a"]).Error; err != nil {
		t.Fatalf("failed to delete the item: %v", err)
	}

	if err := DeleteLabel(db, testUserID, label.ID); err != nil {
		t.Fatalf("expected the label to be considered unused but got %v", err)
	}

	var remaining int64
	if err := db.Table("item_labels").Where("label_id = ?", label.ID).Count(&remaining).Error; err != nil {
		t.Fatalf("failed to count the join rows: %v", err)
	}
	if remaining != 0 {
		t.Errorf("expected the join rows to be swept but %d remained", remaining)
	}
}

func TestFindOrCreateLabels(t *testing.T) {
	db := setupTestDB(t)

	existing, err := CreateLabel(db, testUserID, "work")
	if err != nil {
		t.Fatalf("failed to create the label: %v", err)
	}

	labels, err := FindOrCreateLabels(db, testUserID, []string{"Work", "  work ", "home", "", "   "})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if len(labels) != 2 {
		t.Fatalf("expected 2 labels but got %d (%v)", len(labels), labels)
	}
	if labels[0].ID != existing.ID {
		t.Errorf("expected the existing label %d to be reused but got %d", existing.ID, labels[0].ID)
	}

	all, err := ListLabels(db, testUserID)
	if err != nil {
		t.Fatalf("failed to list the labels: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 labels in total but got %d", len(all))
	}
}

func TestUpdateItemLabels(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a")

	if _, err := UpdateItemLabels(db, testUserID, ids["a"], []string{"Work", "urgent"}, nil); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if names := labelNames(t, db, ids["a"]); !containsAll(names, "work", "urgent") {
		t.Errorf("expected [work urgent] but got %v", names)
	}

	// Adding a label the item already carries must not duplicate it.
	if _, err := UpdateItemLabels(db, testUserID, ids["a"], []string{"work"}, nil); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if names := labelNames(t, db, ids["a"]); !containsAll(names, "work", "urgent") {
		t.Errorf("expected [work urgent] but got %v", names)
	}

	// Removing a name that is not a known label is a no-op rather than an
	// error, and must not create the label just to detach it.
	if _, err := UpdateItemLabels(db, testUserID, ids["a"], nil, []string{"nonexistent"}); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	labels, err := ListLabels(db, testUserID)
	if err != nil {
		t.Fatalf("failed to list the labels: %v", err)
	}
	if len(labels) != 2 {
		t.Errorf("expected removal not to create labels but got %d", len(labels))
	}

	if _, err := UpdateItemLabels(db, testUserID, ids["a"], nil, []string{"URGENT"}); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if names := labelNames(t, db, ids["a"]); !containsAll(names, "work") {
		t.Errorf("expected [work] but got %v", names)
	}

	if _, err := UpdateItemLabels(db, testUserID, 404, []string{"work"}, nil); !errors.Is(err, ErrItemNotFound) {
		t.Errorf("expected %v but got %v", ErrItemNotFound, err)
	}
}

func TestUpdateItemLabelsAddAndRemoveSameName(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a")

	// Removal is applied first, so passing the same name to both leaves the
	// label attached regardless of statement ordering.
	if _, err := UpdateItemLabels(db, testUserID, ids["a"], []string{"work"}, []string{"work"}); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if names := labelNames(t, db, ids["a"]); !containsAll(names, "work") {
		t.Errorf("expected [work] but got %v", names)
	}
}

func TestItemFilterByLabel(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a", "b", "c")

	if _, err := UpdateItemLabels(db, testUserID, ids["a"], []string{"work", "urgent"}, nil); err != nil {
		t.Fatalf("failed to tag: %v", err)
	}
	if _, err := UpdateItemLabels(db, testUserID, ids["b"], []string{"work"}, nil); err != nil {
		t.Fatalf("failed to tag: %v", err)
	}

	tests := []struct {
		name     string
		labels   []string
		expected []string
	}{
		{"no filter", nil, []string{"a", "b", "c"}},
		{"single label", []string{"work"}, []string{"a", "b"}},
		{"single label normalised", []string{"  WORK "}, []string{"a", "b"}},
		{"narrower label", []string{"urgent"}, []string{"a"}},
		{"two labels are combined with AND", []string{"work", "urgent"}, []string{"a"}},
		{"unknown label matches nothing", []string{"nonexistent"}, nil},
		{"unknown label narrows to nothing", []string{"work", "nonexistent"}, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			items, err := ListActive(db, testUserID, ItemFilter{Labels: test.labels})
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

func TestListActivePreloadsLabels(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a")
	if _, err := UpdateItemLabels(db, testUserID, ids["a"], []string{"work"}, nil); err != nil {
		t.Fatalf("failed to tag: %v", err)
	}

	items, err := ListActive(db, testUserID, ItemFilter{})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if len(items) != 1 || len(items[0].Labels) != 1 || items[0].Labels[0].Name != "work" {
		t.Errorf("expected the labels to be preloaded but got %v", items)
	}
}

func TestListCompletedFilterByLabel(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a", "b")

	if _, err := UpdateItemLabels(db, testUserID, ids["a"], []string{"work"}, nil); err != nil {
		t.Fatalf("failed to tag: %v", err)
	}
	for _, title := range []string{"a", "b"} {
		if _, err := SetDone(db, testUserID, ids[title], true); err != nil {
			t.Fatalf("failed to complete %q: %v", title, err)
		}
	}

	items, err := ListCompleted(db, testUserID, ItemFilter{Labels: []string{"work"}})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if len(items) != 1 || items[0].Title != "a" {
		t.Errorf("expected only [a] but got %v", items)
	}
}

func TestFindLabelsByName(t *testing.T) {
	db := setupTestDB(t)
	if _, err := CreateLabel(db, testUserID, "work"); err != nil {
		t.Fatalf("failed to create the label: %v", err)
	}
	if _, err := CreateLabel(db, testUserID, "home"); err != nil {
		t.Fatalf("failed to create the label: %v", err)
	}

	t.Run("empty input returns nothing", func(t *testing.T) {
		labels, err := findLabelsByName(db, testUserID, nil)
		if err != nil {
			t.Fatalf("expected no error but got %v", err)
		}
		if len(labels) != 0 {
			t.Errorf("expected no labels but got %v", labels)
		}
	})

	t.Run("matches existing labels normalising the names", func(t *testing.T) {
		labels, err := findLabelsByName(db, testUserID, []string{"  WORK ", "home"})
		if err != nil {
			t.Fatalf("expected no error but got %v", err)
		}
		if len(labels) != 2 {
			t.Fatalf("expected 2 labels but got %d (%v)", len(labels), labels)
		}
	})

	t.Run("unknown names match nothing", func(t *testing.T) {
		labels, err := findLabelsByName(db, testUserID, []string{"nonexistent"})
		if err != nil {
			t.Fatalf("expected no error but got %v", err)
		}
		if len(labels) != 0 {
			t.Errorf("expected no labels but got %v", labels)
		}
	})
}

func TestFindLabelCrossUserIsNotFound(t *testing.T) {
	db := setupTestDB(t)
	label, err := CreateLabel(db, testUserID, "work")
	if err != nil {
		t.Fatalf("failed to create the label: %v", err)
	}

	// A different user must not be able to load the label; cross-user access
	// is reported as not found rather than leaking existence.
	var loaded Label
	err = findLabel(db, testUserID+1, label.ID, &loaded)
	if !errors.Is(err, ErrLabelNotFound) {
		t.Errorf("expected %v but got %v", ErrLabelNotFound, err)
	}
}
