package database

import (
	"errors"
	"testing"
)

func TestUpdateItem(t *testing.T) {
	db := setupTestDB(t)
	item := Item{Title: "original", Description: "old description", UserID: testUserID}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	updated, err := UpdateItem(db, testUserID, item.ID, "new title", "new description")
	if err != nil {
		t.Fatalf("failed to update item: %v", err)
	}
	if updated.Title != "new title" {
		t.Errorf("expected title %q but got %q", "new title", updated.Title)
	}
	if updated.Description != "new description" {
		t.Errorf("expected description %q but got %q", "new description", updated.Description)
	}

	// Clearing the description writes an empty string (the column is nullable
	// but the convention is to store "" verbatim, matching CreateItem).
	updated, err = UpdateItem(db, testUserID, item.ID, "title only", "")
	if err != nil {
		t.Fatalf("failed to clear description: %v", err)
	}
	if updated.Description != "" {
		t.Errorf("expected empty description but got %q", updated.Description)
	}
}

func TestUpdateItemRequiresTitle(t *testing.T) {
	db := setupTestDB(t)
	item := Item{Title: "original", UserID: testUserID}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	tests := []struct {
		name  string
		title string
	}{
		{"empty", ""},
		{"whitespace only", "   "},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := UpdateItem(db, testUserID, item.ID, test.title, "desc"); err != ErrItemTitleEmpty {
				t.Errorf("expected ErrItemTitleEmpty but got %v", err)
			}
		})
	}
}

func TestUpdateItemTrimsTitle(t *testing.T) {
	db := setupTestDB(t)
	item := Item{Title: "original", UserID: testUserID}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	updated, err := UpdateItem(db, testUserID, item.ID, "  surrounded  ", "desc")
	if err != nil {
		t.Fatalf("failed to update item: %v", err)
	}
	if updated.Title != "surrounded" {
		t.Errorf("expected trimmed title %q but got %q", "surrounded", updated.Title)
	}
}

func TestUpdateItemCrossUserIsNotFound(t *testing.T) {
	db := setupTestDB(t)
	item := Item{Title: "original", UserID: testUserID}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	const otherUserID uint = 999
	_, err := UpdateItem(db, otherUserID, item.ID, "hijacked", "stolen")
	if !errors.Is(err, ErrItemNotFound) {
		t.Errorf("expected ErrItemNotFound but got %v", err)
	}

	// The item is unchanged.
	var got Item
	if err := db.First(&got, item.ID).Error; err != nil {
		t.Fatalf("failed to reload item: %v", err)
	}
	if got.Title != "original" {
		t.Errorf("expected unchanged title %q but got %q", "original", got.Title)
	}
}