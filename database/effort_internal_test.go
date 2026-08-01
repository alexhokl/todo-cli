package database

import (
	"errors"
	"testing"
)

func TestNormaliseEffortName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"high", "high"},
		{"High", "high"},
		{"HIGH", "high"},
		{"  high  ", "high"},
		{"\tHigh\n", "high"},
		{"two words", "two words"},
		{"", ""},
		{"   ", ""},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			if got := NormaliseEffortName(test.input); got != test.expected {
				t.Errorf("expected %q but got %q", test.expected, got)
			}
		})
	}
}

func TestListEfforts(t *testing.T) {
	db := setupTestDB(t)

	// Create out of order so the ordering is exercised.
	for _, name := range []string{"high", "low", "medium"} {
		if _, err := CreateEffort(db, testUserID, name); err != nil {
			t.Fatalf("failed to create %q: %v", name, err)
		}
	}

	efforts, err := ListEfforts(db, testUserID)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	names := make([]string, 0, len(efforts))
	for _, effort := range efforts {
		names = append(names, effort.Name)
	}
	if !equalStrings(names, []string{"high", "low", "medium"}) {
		t.Errorf("expected [high low medium] but got %v", names)
	}
}

func TestCreateEffort(t *testing.T) {
	db := setupTestDB(t)

	effort, err := CreateEffort(db, testUserID, "  High ")
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if effort.Name != "high" {
		t.Errorf("expected the name to be normalised to %q but got %q", "high", effort.Name)
	}
	if effort.UserID != testUserID {
		t.Errorf("expected the user id to be set but got %d", effort.UserID)
	}

	// A duplicate name is reported rather than returning the existing effort.
	if _, err := CreateEffort(db, testUserID, "HIGH"); !errors.Is(err, ErrEffortExists) {
		t.Errorf("expected %v but got %v", ErrEffortExists, err)
	}

	// Two users can each own an effort with the same name.
	if _, err := CreateEffort(db, testUserID+1, "high"); err != nil {
		t.Fatalf("expected no error for a different user but got %v", err)
	}

	// An empty name is rejected.
	if _, err := CreateEffort(db, testUserID, "  "); !errors.Is(err, ErrEffortNameEmpty) {
		t.Errorf("expected %v but got %v", ErrEffortNameEmpty, err)
	}
}

func TestRenameEffort(t *testing.T) {
	db := setupTestDB(t)
	effort, err := CreateEffort(db, testUserID, "low")
	if err != nil {
		t.Fatalf("failed to create the effort: %v", err)
	}

	renamed, err := RenameEffort(db, testUserID, effort.ID, "  Trivial ")
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if renamed.Name != "trivial" {
		t.Errorf("expected the name to be %q but got %q", "trivial", renamed.Name)
	}

	// Renaming to the name it already has is a no-op.
	if _, err := RenameEffort(db, testUserID, effort.ID, "trivial"); err != nil {
		t.Errorf("expected no error renaming to the same name but got %v", err)
	}

	// A collision with another effort is reported.
	other, err := CreateEffort(db, testUserID, "medium")
	if err != nil {
		t.Fatalf("failed to create the other effort: %v", err)
	}
	if _, err := RenameEffort(db, testUserID, effort.ID, "medium"); !errors.Is(err, ErrEffortExists) {
		t.Errorf("expected %v but got %v", ErrEffortExists, err)
	}
	// The other effort must be unchanged.
	var stillOther Effort
	if err := db.First(&stillOther, other.ID).Error; err != nil {
		t.Fatalf("failed to load the other effort: %v", err)
	}
	if stillOther.Name != "medium" {
		t.Errorf("expected the other effort to keep its name but got %q", stillOther.Name)
	}

	// An unknown id is reported.
	if _, err := RenameEffort(db, testUserID, 404, "high"); !errors.Is(err, ErrEffortNotFound) {
		t.Errorf("expected %v but got %v", ErrEffortNotFound, err)
	}

	// A cross-user access is reported as not found.
	if _, err := RenameEffort(db, testUserID+1, effort.ID, "high"); !errors.Is(err, ErrEffortNotFound) {
		t.Errorf("expected %v for a cross-user access but got %v", ErrEffortNotFound, err)
	}

	// An empty name is rejected.
	if _, err := RenameEffort(db, testUserID, effort.ID, "  "); !errors.Is(err, ErrEffortNameEmpty) {
		t.Errorf("expected %v but got %v", ErrEffortNameEmpty, err)
	}
}

func TestDeleteEffort(t *testing.T) {
	db := setupTestDB(t)
	effort, err := CreateEffort(db, testUserID, "low")
	if err != nil {
		t.Fatalf("failed to create the effort: %v", err)
	}

	// An effort in use cannot be deleted.
	item := Item{Title: "blocked", UserID: testUserID, EffortID: &effort.ID}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("failed to create the item: %v", err)
	}
	if err := DeleteEffort(db, testUserID, effort.ID); !errors.Is(err, ErrEffortInUse) {
		t.Errorf("expected %v but got %v", ErrEffortInUse, err)
	}

	// Clearing the item allows the delete.
	if _, err := SetItemEffort(db, testUserID, item.ID, ""); err != nil {
		t.Fatalf("failed to clear the effort: %v", err)
	}
	if err := DeleteEffort(db, testUserID, effort.ID); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	// A second delete is reported as not found.
	if err := DeleteEffort(db, testUserID, effort.ID); !errors.Is(err, ErrEffortNotFound) {
		t.Errorf("expected %v but got %v", ErrEffortNotFound, err)
	}

	// A cross-user access is reported as not found.
	other, err := CreateEffort(db, testUserID, "medium")
	if err != nil {
		t.Fatalf("failed to create the other effort: %v", err)
	}
	if err := DeleteEffort(db, testUserID+1, other.ID); !errors.Is(err, ErrEffortNotFound) {
		t.Errorf("expected %v for a cross-user access but got %v", ErrEffortNotFound, err)
	}
}

func TestFindEffortByName(t *testing.T) {
	db := setupTestDB(t)
	if _, err := CreateEffort(db, testUserID, "high"); err != nil {
		t.Fatalf("failed to create the effort: %v", err)
	}

	effort, err := FindEffortByName(db, testUserID, "  High ")
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if effort.Name != "high" {
		t.Errorf("expected the name to be %q but got %q", "high", effort.Name)
	}

	// An unknown name is reported.
	if _, err := FindEffortByName(db, testUserID, "unknown"); !errors.Is(err, ErrEffortNotFound) {
		t.Errorf("expected %v but got %v", ErrEffortNotFound, err)
	}

	// A cross-user access is reported as not found.
	if _, err := FindEffortByName(db, testUserID+1, "high"); !errors.Is(err, ErrEffortNotFound) {
		t.Errorf("expected %v for a cross-user access but got %v", ErrEffortNotFound, err)
	}

	// An empty name is rejected.
	if _, err := FindEffortByName(db, testUserID, "  "); !errors.Is(err, ErrEffortNameEmpty) {
		t.Errorf("expected %v but got %v", ErrEffortNameEmpty, err)
	}
}

func TestSetItemEffort(t *testing.T) {
	db := setupTestDB(t)
	effort, err := CreateEffort(db, testUserID, "high")
	if err != nil {
		t.Fatalf("failed to create the effort: %v", err)
	}
	item := Item{Title: "task", UserID: testUserID}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("failed to create the item: %v", err)
	}

	// Setting by name attaches the effort.
	updated, err := SetItemEffort(db, testUserID, item.ID, "High")
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if updated.EffortID == nil || *updated.EffortID != effort.ID {
		t.Errorf("expected effort %d but got %v", effort.ID, updated.EffortID)
	}
	if updated.Effort == nil || updated.Effort.Name != "high" {
		t.Errorf("expected the preloaded effort to be high but got %v", updated.Effort)
	}

	// Clearing with an empty name detaches the effort.
	cleared, err := SetItemEffort(db, testUserID, item.ID, "  ")
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if cleared.EffortID != nil {
		t.Errorf("expected the effort to be cleared but got %v", cleared.EffortID)
	}

	// An unknown effort name is reported rather than being created.
	if _, err := SetItemEffort(db, testUserID, item.ID, "unknown"); !errors.Is(err, ErrEffortNotFound) {
		t.Errorf("expected %v but got %v", ErrEffortNotFound, err)
	}

	// An unknown item is reported.
	if _, err := SetItemEffort(db, testUserID, 404, "high"); !errors.Is(err, ErrItemNotFound) {
		t.Errorf("expected %v but got %v", ErrItemNotFound, err)
	}

	// A cross-user access is reported as not found.
	if _, err := SetItemEffort(db, testUserID+1, item.ID, "high"); !errors.Is(err, ErrItemNotFound) {
		t.Errorf("expected %v for a cross-user access but got %v", ErrItemNotFound, err)
	}
}