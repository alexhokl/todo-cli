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

func TestDeleteItem(t *testing.T) {
	db := setupTestDB(t)
	item := Item{Title: "untriaged", UserID: testUserID}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	if err := DeleteItem(db, testUserID, item.ID); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	if _, err := GetItem(db, testUserID, item.ID); !errors.Is(err, ErrItemNotFound) {
		t.Errorf("expected ErrItemNotFound after deletion but got %v", err)
	}

	// A second delete of the same id is also not found.
	if err := DeleteItem(db, testUserID, item.ID); !errors.Is(err, ErrItemNotFound) {
		t.Errorf("expected ErrItemNotFound on second delete but got %v", err)
	}
}

func TestDeleteItemRemovesBlockersAndComments(t *testing.T) {
	db := setupTestDB(t)
	item := Item{Title: "untriaged", UserID: testUserID}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("failed to create item: %v", err)
	}
	if _, err := CreateBlocker(db, testUserID, item.ID, "waiting on review"); err != nil {
		t.Fatalf("failed to create blocker: %v", err)
	}
	if _, err := CreateComment(db, testUserID, item.ID, "drafted a reply"); err != nil {
		t.Fatalf("failed to create comment: %v", err)
	}

	if err := DeleteItem(db, testUserID, item.ID); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	var blockerCount int64
	db.Model(&Blocker{}).Where("item_id = ?", item.ID).Count(&blockerCount)
	if blockerCount != 0 {
		t.Errorf("expected no blockers but got %d", blockerCount)
	}
	var commentCount int64
	db.Model(&Comment{}).Where("item_id = ?", item.ID).Count(&commentCount)
	if commentCount != 0 {
		t.Errorf("expected no comments but got %d", commentCount)
	}
}

func TestDeleteItemRejectsLinkedItems(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a", "b")
	// `a` is triaged by seedItems; untriage it by reopening via SetDone to
	// drop its priority, then link it to `b`.
	if _, err := SetDone(db, testUserID, ids["a"], true); err != nil {
		t.Fatalf("failed to complete a: %v", err)
	}
	if _, err := SetDone(db, testUserID, ids["a"], false); err != nil {
		t.Fatalf("failed to reopen a: %v", err)
	}
	if _, err := UpdateItemLinks(db, testUserID, ids["a"], []uint{ids["b"]}, nil); err != nil {
		t.Fatalf("failed to link a to b: %v", err)
	}

	if err := DeleteItem(db, testUserID, ids["a"]); !errors.Is(err, ErrItemHasLinks) {
		t.Fatalf("expected ErrItemHasLinks but got %v", err)
	}

	// The item is still present.
	if _, err := GetItem(db, testUserID, ids["a"]); err != nil {
		t.Errorf("expected the item to still exist but got %v", err)
	}
}

func TestDeleteItemRejectsTriaged(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a") // seedItems triages via MoveItem(bottom)

	if err := DeleteItem(db, testUserID, ids["a"]); !errors.Is(err, ErrItemNotUntriaged) {
		t.Errorf("expected ErrItemNotUntriaged but got %v", err)
	}
}

func TestDeleteItemRejectsCompleted(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a")
	if _, err := SetDone(db, testUserID, ids["a"], true); err != nil {
		t.Fatalf("failed to complete a: %v", err)
	}

	if err := DeleteItem(db, testUserID, ids["a"]); !errors.Is(err, ErrItemNotUntriaged) {
		t.Errorf("expected ErrItemNotUntriaged but got %v", err)
	}
}

func TestDeleteItemRejectsCrossUser(t *testing.T) {
	db := setupTestDB(t)
	item := Item{Title: "untriaged", UserID: testUserID}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	const otherUserID uint = 999
	if err := DeleteItem(db, otherUserID, item.ID); !errors.Is(err, ErrItemNotFound) {
		t.Errorf("expected ErrItemNotFound but got %v", err)
	}

	// The item is unchanged.
	var got Item
	if err := db.First(&got, item.ID).Error; err != nil {
		t.Fatalf("failed to reload item: %v", err)
	}
	if got.Title != "untriaged" {
		t.Errorf("expected unchanged title %q but got %q", "untriaged", got.Title)
	}
}

func TestDeleteItemCleansUpJoinRows(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a", "b")
	// Untriage `a` so it becomes deletable, then link a<->b.
	if _, err := SetDone(db, testUserID, ids["a"], true); err != nil {
		t.Fatalf("failed to complete a: %v", err)
	}
	if _, err := SetDone(db, testUserID, ids["a"], false); err != nil {
		t.Fatalf("failed to reopen a: %v", err)
	}
	if _, err := UpdateItemLinks(db, testUserID, ids["a"], []uint{ids["b"]}, nil); err != nil {
		t.Fatalf("failed to link a to b: %v", err)
	}
	// Detach the link so `a` becomes deletable.
	if _, err := UpdateItemLinks(db, testUserID, ids["a"], nil, []uint{ids["b"]}); err != nil {
		t.Fatalf("failed to unlink a from b: %v", err)
	}

	if err := DeleteItem(db, testUserID, ids["a"]); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	// No join rows should reference the deleted item in either direction.
	var joinCount int64
	db.Model(&itemLinks{}).Where("item_id = ? OR linked_item_id = ?", ids["a"], ids["a"]).Count(&joinCount)
	if joinCount != 0 {
		t.Errorf("expected no join rows referencing the deleted item but got %d", joinCount)
	}
}

func TestDeleteItemSoftDeletes(t *testing.T) {
	db := setupTestDB(t)
	item := Item{Title: "untriaged", UserID: testUserID}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	if err := DeleteItem(db, testUserID, item.ID); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	// The row still exists but is soft-deleted (deleted_at IS NOT NULL).
	var got Item
	if err := db.Unscoped().First(&got, item.ID).Error; err != nil {
		t.Fatalf("failed to reload soft-deleted item: %v", err)
	}
	if got.DeletedAt.Valid != true {
		t.Errorf("expected the item to be soft-deleted but deleted_at is null")
	}
}