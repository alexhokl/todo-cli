package database

import (
	"fmt"

	"gorm.io/gorm"
)

// AutoMigrate creates or updates the database schema for all known models.
// Application invariants that the ORM cannot express as struct tags are
// enforced via SQLite triggers, since SQLite does not support adding CHECK
// constraints to an existing table without a full table rebuild.
func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&User{},
		&TailscaleAddress{},
		&List{},
		&Label{},
		&Effort{},
		&Blocker{},
		&Comment{},
		&Item{},
	); err != nil {
		return err
	}

	// An item has a priority exactly when it is active and triaged, so a done
	// item must never carry one. Enforcing it at the database layer prevents
	// a stray UPDATE (or a future code path that bypasses SetDone) from
	// silently leaking a completed row into the active listing. SQLite has no
	// ALTER TABLE ADD CONSTRAINT, so a trigger raises ABORT on the offending
	// INSERT or UPDATE. CREATE TRIGGER IF NOT EXISTS keeps this idempotent.
	const condition = "NEW.done = 1 AND NEW.priority IS NOT NULL"
	const message = "a completed item must not carry a priority"
	for _, event := range []string{"INSERT", "UPDATE"} {
		stmt := fmt.Sprintf(
			"CREATE TRIGGER IF NOT EXISTS items_priority_inactive_on_%s "+
				"BEFORE %s ON items "+
				"WHEN %s "+
				"BEGIN SELECT RAISE(ABORT, '%s'); END",
			event, event, condition, message,
		)
		if err := db.Exec(stmt).Error; err != nil {
			return fmt.Errorf("failed to create items_priority_inactive_on_%s trigger: %w", event, err)
		}
	}

	return nil
}
