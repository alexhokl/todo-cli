package database

import (
	"errors"
	"testing"
)

func TestListBlockers(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a")

	// An item with no blockers yields an empty list.
	blockers, err := ListBlockers(db, testUserID, ids["a"])
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if len(blockers) != 0 {
		t.Errorf("expected no blockers but got %v", blockers)
	}

	// Create out of order so the id ordering is exercised.
	if _, err := CreateBlocker(db, testUserID, ids["a"], "waiting on legal"); err != nil {
		t.Fatalf("failed to create the blocker: %v", err)
	}
	if _, err := CreateBlocker(db, testUserID, ids["a"], "blocked by upstream"); err != nil {
		t.Fatalf("failed to create the blocker: %v", err)
	}

	blockers, err = ListBlockers(db, testUserID, ids["a"])
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if len(blockers) != 2 {
		t.Fatalf("expected 2 blockers but got %d", len(blockers))
	}
	if blockers[0].Description != "waiting on legal" {
		t.Errorf("expected the first blocker by id but got %q", blockers[0].Description)
	}
}

func TestCreateBlocker(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a")

	blocker, err := CreateBlocker(db, testUserID, ids["a"], "  needs review  ")
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if blocker.Description != "needs review" {
		t.Errorf("expected the description to be trimmed to %q but got %q", "needs review", blocker.Description)
	}
	if blocker.ItemID != ids["a"] {
		t.Errorf("expected the item id to be %d but got %d", ids["a"], blocker.ItemID)
	}
	if blocker.UserID != testUserID {
		t.Errorf("expected the user id to be %d but got %d", testUserID, blocker.UserID)
	}

	// An empty description is rejected.
	if _, err := CreateBlocker(db, testUserID, ids["a"], "   "); !errors.Is(err, ErrBlockerDescriptionEmpty) {
		t.Errorf("expected %v but got %v", ErrBlockerDescriptionEmpty, err)
	}

	// An unknown item is reported.
	if _, err := CreateBlocker(db, testUserID, 404, "blocked"); !errors.Is(err, ErrItemNotFound) {
		t.Errorf("expected %v but got %v", ErrItemNotFound, err)
	}

	// A cross-user access is reported as not found.
	if _, err := CreateBlocker(db, testUserID+1, ids["a"], "blocked"); !errors.Is(err, ErrItemNotFound) {
		t.Errorf("expected %v for a cross-user access but got %v", ErrItemNotFound, err)
	}
}

func TestUpdateBlocker(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a")
	blocker, err := CreateBlocker(db, testUserID, ids["a"], "old reason")
	if err != nil {
		t.Fatalf("failed to create the blocker: %v", err)
	}

	updated, err := UpdateBlocker(db, testUserID, blocker.ID, "  new reason  ")
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if updated.Description != "new reason" {
		t.Errorf("expected the description to be %q but got %q", "new reason", updated.Description)
	}

	// An empty description is rejected.
	if _, err := UpdateBlocker(db, testUserID, blocker.ID, "  "); !errors.Is(err, ErrBlockerDescriptionEmpty) {
		t.Errorf("expected %v but got %v", ErrBlockerDescriptionEmpty, err)
	}

	// An unknown id is reported.
	if _, err := UpdateBlocker(db, testUserID, 404, "anything"); !errors.Is(err, ErrBlockerNotFound) {
		t.Errorf("expected %v but got %v", ErrBlockerNotFound, err)
	}

	// A cross-user access is reported as not found.
	if _, err := UpdateBlocker(db, testUserID+1, blocker.ID, "anything"); !errors.Is(err, ErrBlockerNotFound) {
		t.Errorf("expected %v for a cross-user access but got %v", ErrBlockerNotFound, err)
	}
}

func TestDeleteBlocker(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a")
	blocker, err := CreateBlocker(db, testUserID, ids["a"], "blocked")
	if err != nil {
		t.Fatalf("failed to create the blocker: %v", err)
	}

	if err := DeleteBlocker(db, testUserID, blocker.ID); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	// A second delete is reported as not found.
	if err := DeleteBlocker(db, testUserID, blocker.ID); !errors.Is(err, ErrBlockerNotFound) {
		t.Errorf("expected %v but got %v", ErrBlockerNotFound, err)
	}

	// A cross-user access is reported as not found.
	other, err := CreateBlocker(db, testUserID, ids["a"], "still blocked")
	if err != nil {
		t.Fatalf("failed to create the other blocker: %v", err)
	}
	if err := DeleteBlocker(db, testUserID+1, other.ID); !errors.Is(err, ErrBlockerNotFound) {
		t.Errorf("expected %v for a cross-user access but got %v", ErrBlockerNotFound, err)
	}
}

func TestListBlockersPreloadsOnFindItem(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a")
	if _, err := CreateBlocker(db, testUserID, ids["a"], "waiting on legal"); err != nil {
		t.Fatalf("failed to create the blocker: %v", err)
	}

	var item Item
	if err := findItem(db, testUserID, ids["a"], &item); err != nil {
		t.Fatalf("failed to load the item: %v", err)
	}
	if len(item.Blockers) != 1 || item.Blockers[0].Description != "waiting on legal" {
		t.Errorf("expected the blocker to be preloaded but got %v", item.Blockers)
	}
}