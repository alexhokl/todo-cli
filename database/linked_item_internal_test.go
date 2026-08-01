package database

import (
	"errors"
	"testing"
)

func TestUpdateItemLinksSymmetric(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a", "b", "c")

	// Linking a to b should also link b to a.
	if _, err := UpdateItemLinks(db, testUserID, ids["a"], []uint{ids["b"]}, nil); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	// a carries b, and b carries a (symmetric).
	var a, b Item
	if err := findItem(db, testUserID, ids["a"], &a); err != nil {
		t.Fatalf("failed to load a: %v", err)
	}
	if err := findItem(db, testUserID, ids["b"], &b); err != nil {
		t.Fatalf("failed to load b: %v", err)
	}
	if len(a.LinkedItems) != 1 || a.LinkedItems[0].ID != ids["b"] {
		t.Errorf("expected a to link to b but got %v", a.LinkedItems)
	}
	if len(b.LinkedItems) != 1 || b.LinkedItems[0].ID != ids["a"] {
		t.Errorf("expected b to link to a (symmetric) but got %v", b.LinkedItems)
	}
	// c is unlinked.
	var c Item
	if err := findItem(db, testUserID, ids["c"], &c); err != nil {
		t.Fatalf("failed to load c: %v", err)
	}
	if len(c.LinkedItems) != 0 {
		t.Errorf("expected c to carry no links but got %v", c.LinkedItems)
	}
}

func TestUpdateItemLinksRemovesBothDirections(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a", "b")

	if _, err := UpdateItemLinks(db, testUserID, ids["a"], []uint{ids["b"]}, nil); err != nil {
		t.Fatalf("failed to link a to b: %v", err)
	}

	// Unlinking from a removes the link on b as well.
	if _, err := UpdateItemLinks(db, testUserID, ids["a"], nil, []uint{ids["b"]}); err != nil {
		t.Fatalf("failed to unlink: %v", err)
	}

	var a, b Item
	if err := findItem(db, testUserID, ids["a"], &a); err != nil {
		t.Fatalf("failed to load a: %v", err)
	}
	if err := findItem(db, testUserID, ids["b"], &b); err != nil {
		t.Fatalf("failed to load b: %v", err)
	}
	if len(a.LinkedItems) != 0 {
		t.Errorf("expected a to carry no links but got %v", a.LinkedItems)
	}
	if len(b.LinkedItems) != 0 {
		t.Errorf("expected b to carry no links (symmetric removal) but got %v", b.LinkedItems)
	}
}

func TestUpdateItemLinksIdempotent(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a", "b")

	if _, err := UpdateItemLinks(db, testUserID, ids["a"], []uint{ids["b"]}, nil); err != nil {
		t.Fatalf("first link failed: %v", err)
	}
	// Linking again is a no-op, not a duplicate or an error.
	if _, err := UpdateItemLinks(db, testUserID, ids["a"], []uint{ids["b"]}, nil); err != nil {
		t.Fatalf("second link failed: %v", err)
	}

	var a Item
	if err := findItem(db, testUserID, ids["a"], &a); err != nil {
		t.Fatalf("failed to load a: %v", err)
	}
	if len(a.LinkedItems) != 1 {
		t.Errorf("expected exactly one link after idempotent re-add but got %v", a.LinkedItems)
	}
}

func TestUpdateItemLinksAddAndRemoveSameID(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a", "b")

	if _, err := UpdateItemLinks(db, testUserID, ids["a"], []uint{ids["b"]}, nil); err != nil {
		t.Fatalf("failed to link: %v", err)
	}

	// Removal runs first, so passing the same id to both leaves it attached.
	if _, err := UpdateItemLinks(db, testUserID, ids["a"], []uint{ids["b"]}, []uint{ids["b"]}); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	var a Item
	if err := findItem(db, testUserID, ids["a"], &a); err != nil {
		t.Fatalf("failed to load a: %v", err)
	}
	if len(a.LinkedItems) != 1 {
		t.Errorf("expected the link to remain (remove-then-add) but got %v", a.LinkedItems)
	}
}

func TestUpdateItemLinksRejectsSelfLink(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a")

	if _, err := UpdateItemLinks(db, testUserID, ids["a"], []uint{ids["a"]}, nil); !errors.Is(err, ErrItemLinkToSelf) {
		t.Errorf("expected %v but got %v", ErrItemLinkToSelf, err)
	}
	if _, err := UpdateItemLinks(db, testUserID, ids["a"], nil, []uint{ids["a"]}); !errors.Is(err, ErrItemLinkToSelf) {
		t.Errorf("expected %v on self-unlink but got %v", ErrItemLinkToSelf, err)
	}
}

func TestUpdateItemLinksRejectsUnknownTarget(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a")

	if _, err := UpdateItemLinks(db, testUserID, ids["a"], []uint{404}, nil); !errors.Is(err, ErrItemNotFound) {
		t.Errorf("expected %v but got %v", ErrItemNotFound, err)
	}
}

func TestUpdateItemLinksRejectsCrossUserTarget(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a")

	// Create an item owned by another user.
	other := Item{Title: "theirs", UserID: testUserID + 1}
	if err := db.Create(&other).Error; err != nil {
		t.Fatalf("failed to create the other item: %v", err)
	}

	if _, err := UpdateItemLinks(db, testUserID, ids["a"], []uint{other.ID}, nil); !errors.Is(err, ErrItemNotFound) {
		t.Errorf("expected %v for a cross-user target but got %v", ErrItemNotFound, err)
	}
}

func TestUpdateItemLinksFiltersSoftDeletedTargets(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a", "b")

	if _, err := UpdateItemLinks(db, testUserID, ids["a"], []uint{ids["b"]}, nil); err != nil {
		t.Fatalf("failed to link: %v", err)
	}

	// Soft-delete b (the linked item).
	if _, err := SetDone(db, testUserID, ids["b"], true); err != nil {
		t.Fatalf("failed to complete b: %v", err)
	}
	if err := db.Delete(&Item{}, ids["b"]).Error; err != nil {
		t.Fatalf("failed to soft-delete b: %v", err)
	}

	// a's preloaded LinkedItems must not include the soft-deleted b.
	var a Item
	if err := findItem(db, testUserID, ids["a"], &a); err != nil {
		t.Fatalf("failed to load a: %v", err)
	}
	if len(a.LinkedItems) != 0 {
		t.Errorf("expected the soft-deleted link to be filtered out but got %v", a.LinkedItems)
	}
}