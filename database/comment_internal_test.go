package database

import (
	"errors"
	"testing"
)

func TestListComments(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a")

	// An item with no comments yields an empty list.
	comments, err := ListComments(db, testUserID, ids["a"])
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if len(comments) != 0 {
		t.Errorf("expected no comments but got %v", comments)
	}

	// Create out of order so the id ordering is exercised.
	if _, err := CreateComment(db, testUserID, ids["a"], "first remark"); err != nil {
		t.Fatalf("failed to create the comment: %v", err)
	}
	if _, err := CreateComment(db, testUserID, ids["a"], "second remark"); err != nil {
		t.Fatalf("failed to create the comment: %v", err)
	}

	comments, err = ListComments(db, testUserID, ids["a"])
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if len(comments) != 2 {
		t.Fatalf("expected 2 comments but got %d", len(comments))
	}
	if comments[0].Body != "first remark" {
		t.Errorf("expected the first comment by id but got %q", comments[0].Body)
	}
	// The author is resolved from the preloaded User.
	if comments[0].User.Username != "testuser" {
		t.Errorf("expected the author to be testuser but got %q", comments[0].User.Username)
	}
}

func TestCreateComment(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a")

	comment, err := CreateComment(db, testUserID, ids["a"], "  needs review  ")
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if comment.Body != "needs review" {
		t.Errorf("expected the body to be trimmed to %q but got %q", "needs review", comment.Body)
	}
	if comment.ItemID != ids["a"] {
		t.Errorf("expected the item id to be %d but got %d", ids["a"], comment.ItemID)
	}
	if comment.UserID != testUserID {
		t.Errorf("expected the user id to be %d but got %d", testUserID, comment.UserID)
	}
	// The author is resolved from the preloaded User.
	if comment.User.Username != "testuser" {
		t.Errorf("expected the author to be testuser but got %q", comment.User.Username)
	}

	// An empty body is rejected.
	if _, err := CreateComment(db, testUserID, ids["a"], "   "); !errors.Is(err, ErrCommentBodyEmpty) {
		t.Errorf("expected %v but got %v", ErrCommentBodyEmpty, err)
	}

	// An unknown item is reported.
	if _, err := CreateComment(db, testUserID, 404, "remark"); !errors.Is(err, ErrItemNotFound) {
		t.Errorf("expected %v but got %v", ErrItemNotFound, err)
	}

	// A cross-user access is reported as not found.
	if _, err := CreateComment(db, testUserID+1, ids["a"], "remark"); !errors.Is(err, ErrItemNotFound) {
		t.Errorf("expected %v for a cross-user access but got %v", ErrItemNotFound, err)
	}
}

func TestUpdateComment(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a")
	comment, err := CreateComment(db, testUserID, ids["a"], "old remark")
	if err != nil {
		t.Fatalf("failed to create the comment: %v", err)
	}

	updated, err := UpdateComment(db, testUserID, comment.ID, "  new remark  ")
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if updated.Body != "new remark" {
		t.Errorf("expected the body to be %q but got %q", "new remark", updated.Body)
	}
	// The author is still resolved after the re-read.
	if updated.User.Username != "testuser" {
		t.Errorf("expected the author to be testuser but got %q", updated.User.Username)
	}

	// An empty body is rejected.
	if _, err := UpdateComment(db, testUserID, comment.ID, "  "); !errors.Is(err, ErrCommentBodyEmpty) {
		t.Errorf("expected %v but got %v", ErrCommentBodyEmpty, err)
	}

	// An unknown id is reported.
	if _, err := UpdateComment(db, testUserID, 404, "anything"); !errors.Is(err, ErrCommentNotFound) {
		t.Errorf("expected %v but got %v", ErrCommentNotFound, err)
	}

	// A cross-user access is reported as not found.
	if _, err := UpdateComment(db, testUserID+1, comment.ID, "anything"); !errors.Is(err, ErrCommentNotFound) {
		t.Errorf("expected %v for a cross-user access but got %v", ErrCommentNotFound, err)
	}
}

func TestDeleteComment(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a")
	comment, err := CreateComment(db, testUserID, ids["a"], "remark")
	if err != nil {
		t.Fatalf("failed to create the comment: %v", err)
	}

	if err := DeleteComment(db, testUserID, comment.ID); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	// A second delete is reported as not found.
	if err := DeleteComment(db, testUserID, comment.ID); !errors.Is(err, ErrCommentNotFound) {
		t.Errorf("expected %v but got %v", ErrCommentNotFound, err)
	}

	// A cross-user access is reported as not found.
	other, err := CreateComment(db, testUserID, ids["a"], "still remarked")
	if err != nil {
		t.Fatalf("failed to create the other comment: %v", err)
	}
	if err := DeleteComment(db, testUserID+1, other.ID); !errors.Is(err, ErrCommentNotFound) {
		t.Errorf("expected %v for a cross-user access but got %v", ErrCommentNotFound, err)
	}
}

func TestListCommentsPreloadsOnFindItem(t *testing.T) {
	db := setupTestDB(t)
	ids := seedItems(t, db, "a")
	if _, err := CreateComment(db, testUserID, ids["a"], "remark"); err != nil {
		t.Fatalf("failed to create the comment: %v", err)
	}

	var item Item
	if err := findItem(db, testUserID, ids["a"], &item); err != nil {
		t.Fatalf("failed to load the item: %v", err)
	}
	if len(item.Comments) != 1 || item.Comments[0].Body != "remark" {
		t.Errorf("expected the comment to be preloaded but got %v", item.Comments)
	}
}