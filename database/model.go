package database

import (
	"time"

	"gorm.io/gorm"
)

// User is an authenticated Tailscale user. It is created on first sight by
// the Tailscale authentication interceptor and referenced by TailscaleAddress
// and the per-user records (List, Label, Effort, Item).
type User struct {
	gorm.Model
	Username string `gorm:"not null;unique"`
}

// TailscaleAddress caches the mapping from a Tailscale peer IP address to a
// User, so the authentication interceptor can resolve subsequent requests from
// the same address without calling WhoIs again.
type TailscaleAddress struct {
	Address string `gorm:"primaryKey;not null;unique"`
	UserID  uint   `gorm:"not null"`
	User    User   `gorm:"foreignKey:UserID"`
}

// List is a named grouping of todo items.
type List struct {
	gorm.Model
	Name   string `gorm:"not null;uniqueIndex:idx_list_user,priority:1"`
	UserID uint   `gorm:"not null;uniqueIndex:idx_list_user,priority:2;index"`
	User   User   `gorm:"foreignKey:UserID"`
}

// Label is a tag that can be attached to any number of items.
type Label struct {
	gorm.Model
	// Name is stored in its normalised form: trimmed and lower cased, so that
	// "Work", " work " and "WORK" all resolve to the same label. Use
	// NormaliseLabelName before comparing against this column.
	Name   string `gorm:"not null;uniqueIndex:idx_label_user,priority:1"`
	UserID uint   `gorm:"not null;uniqueIndex:idx_label_user,priority:2;index"`
	User   User   `gorm:"foreignKey:UserID"`
}

// Effort is a per-user named level of effort (e.g. "low", "medium", "high").
// An item carries at most one effort via Item.EffortID, so unlike labels there
// is no join table and deleting an effort requires that no item references it.
type Effort struct {
	gorm.Model
	// Name is stored in its normalised form: trimmed and lower cased, so that
	// "High", " high " and "HIGH" all resolve to the same effort. Use
	// NormaliseEffortName before comparing against this column.
	Name   string `gorm:"not null;uniqueIndex:idx_effort_user,priority:1"`
	UserID uint   `gorm:"not null;uniqueIndex:idx_effort_user,priority:2;index"`
	User   User   `gorm:"foreignKey:UserID"`
}

// Item is a single todo item, optionally belonging to a List.
type Item struct {
	gorm.Model
	Title       string `gorm:"not null"`
	Description string `gorm:""`
	Done        bool   `gorm:"not null;default:false;index:idx_items_order,priority:1"`
	DueDate     *time.Time
	// Priority is the sparse fractional rank used for manual ordering of
	// active, triaged items. It is non-nil exactly when Done is false and the
	// item has been triaged. Larger values sort first. Gaps between adjacent
	// values are intentional: inserting between two neighbours takes their
	// midpoint, so a move rewrites a single row. Items are created untriaged
	// (Priority is nil); triage happens via MoveItem with top/bottom or a
	// relative anchor.
	Priority *float64 `gorm:"column:priority;index:idx_items_order,priority:2"`
	ListID   *uint
	List     *List `gorm:"foreignKey:ListID"`
	UserID   uint  `gorm:"not null;index"`
	User     User  `gorm:"foreignKey:UserID"`
	// Labels are the tags attached to this item. The join table carries no
	// soft delete column, so rows survive a soft deleted item; DeleteLabel
	// sweeps any that are left behind.
	Labels []Label `gorm:"many2many:item_labels;"`
	// EffortID is the optional single effort level attached to this item.
	// Unlike labels, an item carries at most one effort, so the relationship is
	// a belongs-to rather than a many-to-many and there is no join table.
	EffortID *uint
	Effort   *Effort `gorm:"foreignKey:EffortID"`
}