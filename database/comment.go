package database

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

var (
	// ErrCommentNotFound is returned when a comment does not exist.
	ErrCommentNotFound = errors.New("comment not found")
	// ErrCommentBodyEmpty is returned when a comment body is empty after
	// trimming surrounding whitespace.
	ErrCommentBodyEmpty = errors.New("comment body must not be empty")
)

// ListComments returns every comment attached to the given item, ordered by
// identifier (creation order). The item must exist and belong to the caller;
// cross-user access is reported as not found rather than leaking existence.
// The author (User) is preloaded so the caller can render the username.
func ListComments(db *gorm.DB, userID uint, itemID uint) ([]Comment, error) {
	var item Item
	if err := findItem(db, userID, itemID, &item); err != nil {
		return nil, err
	}

	var comments []Comment
	if err := db.Preload("User").Where("item_id = ?", itemID).Where("user_id = ?", userID).
		Order("id ASC").Find(&comments).Error; err != nil {
		return nil, fmt.Errorf("failed to list the comments: %w", err)
	}

	return comments, nil
}

// CreateComment attaches a new comment to an item. The item must exist and
// belong to the caller; the body must be non-empty after trimming. The author
// is recorded as the caller via the denormalised UserID.
func CreateComment(db *gorm.DB, userID uint, itemID uint, body string) (*Comment, error) {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return nil, ErrCommentBodyEmpty
	}

	var item Item
	var comment Comment
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := findItem(tx, userID, itemID, &item); err != nil {
			return err
		}

		comment = Comment{
			Body:   trimmed,
			ItemID: itemID,
			UserID: userID,
		}
		if err := tx.Create(&comment).Error; err != nil {
			return fmt.Errorf("failed to create the comment: %w", err)
		}

		// Re-read with the User preloaded so the caller can render the author
		// without a second fetch.
		return findComment(tx, userID, comment.ID, &comment)
	})
	if err != nil {
		return nil, err
	}

	return &comment, nil
}

// UpdateComment edits the body of an existing comment. The comment must exist
// and belong to the caller; the body must be non-empty after trimming.
func UpdateComment(db *gorm.DB, userID uint, commentID uint, body string) (*Comment, error) {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return nil, ErrCommentBodyEmpty
	}

	var comment Comment
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := findComment(tx, userID, commentID, &comment); err != nil {
			return err
		}

		// Model(&Comment{}) with an explicit Where avoids the GORM stale-struct
		// gotcha where Model(&instance) uses the instance's field values.
		if err := tx.Model(&Comment{}).Where("id = ?", commentID).
			UpdateColumn("body", trimmed).Error; err != nil {
			return fmt.Errorf("failed to update the comment: %w", err)
		}

		// Re-read into a fresh struct so preloaded fields do not retain stale
		// values from before the update.
		comment = Comment{}
		return findComment(tx, userID, commentID, &comment)
	})
	if err != nil {
		return nil, err
	}

	return &comment, nil
}

// DeleteComment removes a comment. The comment must exist and belong to the
// caller; cross-user access is reported as not found.
func DeleteComment(db *gorm.DB, userID uint, commentID uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var comment Comment
		if err := findComment(tx, userID, commentID, &comment); err != nil {
			return err
		}

		if err := tx.Delete(&comment).Error; err != nil {
			return fmt.Errorf("failed to delete the comment: %w", err)
		}

		return nil
	})
}

// findComment loads a comment by identifier, translating a missing row into
// ErrCommentNotFound. The query is scoped to the given user so cross-user
// access is reported as not found rather than leaking existence. The User is
// preloaded so the author can be rendered without a second fetch.
func findComment(tx *gorm.DB, userID uint, id uint, comment *Comment) error {
	err := tx.Preload("User").Where("user_id = ?", userID).First(comment, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("%w: %d", ErrCommentNotFound, id)
	}
	if err != nil {
		return fmt.Errorf("failed to query the comment: %w", err)
	}

	return nil
}