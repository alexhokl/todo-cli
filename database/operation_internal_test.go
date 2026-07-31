package database

import (
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// testUserID is the identifier of the user seeded by setupTestDB. It matches the
// fixed identifier injected by the dummy authentication interceptor, so tests
// exercise the same code path as production.
const testUserID uint = 1

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := AutoMigrate(db); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}
	if err := db.Create(&User{Username: "testuser"}).Error; err != nil {
		t.Fatalf("failed to seed the test user: %v", err)
	}
	return db
}

func TestAutoMigrate(t *testing.T) {
	db := setupTestDB(t)

	tests := []struct {
		name  string
		model any
	}{
		{"List", &List{}},
		{"Item", &Item{}},
		{"User", &User{}},
		{"TailscaleAddress", &TailscaleAddress{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !db.Migrator().HasTable(test.model) {
				t.Errorf("expected table for %s to exist", test.name)
			}
		})
	}
}

func TestCreateItem(t *testing.T) {
	db := setupTestDB(t)

	item := Item{Title: "write tests", UserID: testUserID}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if item.ID == 0 {
		t.Errorf("expected an assigned ID, got %d", item.ID)
	}
	if item.Done {
		t.Errorf("expected new item to not be done")
	}
}

func TestOpen(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{"empty path", "", true},
		{"nested path", filepath.Join(t.TempDir(), "nested", "todo.db"), false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db, err := Open(test.path)
			if test.expectError {
				if err == nil {
					t.Errorf("expected an error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		if !db.Migrator().HasTable(&Item{}) {
			t.Errorf("expected Item table to exist after Open")
		}
		})
	}
}
