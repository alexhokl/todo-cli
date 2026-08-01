package database

import (
	"testing"
	"time"
)

func TestUpdateItemDueDateSetsAndClears(t *testing.T) {
	db := setupTestDB(t)
	item := Item{Title: "item", UserID: testUserID}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	dueDate := time.Date(2026, time.August, 15, 0, 0, 0, 0, time.Local)
	updated, err := UpdateItemDueDate(db, testUserID, item.ID, &dueDate)
	if err != nil {
		t.Fatalf("failed to set due date: %v", err)
	}
	if updated.DueDate == nil || !updated.DueDate.Equal(dueDate) {
		t.Errorf("expected due date %v but got %v", dueDate, updated.DueDate)
	}

	updated, err = UpdateItemDueDate(db, testUserID, item.ID, nil)
	if err != nil {
		t.Fatalf("failed to clear due date: %v", err)
	}
	if updated.DueDate != nil {
		t.Errorf("expected due date to be cleared but got %v", updated.DueDate)
	}
}
