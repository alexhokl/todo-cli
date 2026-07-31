package database

import (
	"time"

	"gorm.io/gorm"
)

// User is an authenticated Tailscale user. It is created on first sight by
// the Tailscale authentication interceptor and referenced by TailscaleAddress
// and the per-user records (List, Label, Todo).
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

// Label is a tag that can be attached to any number of todos.
type Label struct {
	gorm.Model
	// Name is stored in its normalised form: trimmed and lower cased, so that
	// "Work", " work " and "WORK" all resolve to the same label. Use
	// NormaliseLabelName before comparing against this column.
	Name   string `gorm:"not null;uniqueIndex:idx_label_user,priority:1"`
	UserID uint   `gorm:"not null;uniqueIndex:idx_label_user,priority:2;index"`
	User   User   `gorm:"foreignKey:UserID"`
}

// Todo is a single todo item, optionally belonging to a List.
type Todo struct {
	gorm.Model
	Title       string `gorm:"not null"`
	Description string `gorm:""`
	Done        bool   `gorm:"not null;default:false;index:idx_todos_order,priority:1"`
	DueDate     *time.Time
	// Position is the sparse fractional rank used for manual ordering of
	// active todos. It is non-nil exactly when Done is false. Larger values
	// sort later. Gaps between adjacent values are intentional: inserting
	// between two neighbours takes their midpoint, so a move rewrites a
	// single row.
	Position *float64 `gorm:"index:idx_todos_order,priority:2"`
	ListID   *uint
	List     *List `gorm:"foreignKey:ListID"`
	UserID   uint  `gorm:"not null;index"`
	User     User  `gorm:"foreignKey:UserID"`
	// Labels are the tags attached to this todo. The join table carries no
	// soft delete column, so rows survive a soft deleted todo; DeleteLabel
	// sweeps any that are left behind.
	Labels []Label `gorm:"many2many:todo_labels;"`
}