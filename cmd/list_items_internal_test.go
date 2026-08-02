package cmd

import (
	"testing"

	"github.com/alexhokl/todo-cli/proto"
)

// resetListItemsOpts zeroes the package-level options so a test leaves no
// state behind for the next one. Cobra parses flags into the same struct on
// every invocation, so without this a flag set in one test would leak.
func resetListItemsOpts() {
	listItemsOpts = listItemsOptions{}
}

func TestResolveItemView(t *testing.T) {
	tests := []struct {
		name      string
		setup     func()
		expected  proto.ItemView
		wantError bool
	}{
		{
			name:     "no flags defaults to triaged active view",
			setup:    func() {},
			expected: proto.ItemView_ITEM_VIEW_TRIAGED,
		},
		{
			name:     "untriaged flag",
			setup:    func() { listItemsOpts.Untriaged = true },
			expected: proto.ItemView_ITEM_VIEW_UNTRIAGED,
		},
		{
			name:     "time-sensitive flag",
			setup:    func() { listItemsOpts.TimeSensitive = true },
			expected: proto.ItemView_ITEM_VIEW_TIME_SENSITIVE,
		},
		{
			name:     "done flag",
			setup:    func() { listItemsOpts.Done = true },
			expected: proto.ItemView_ITEM_VIEW_DONE,
		},
		{
			name:      "untriaged and time-sensitive are mutually exclusive",
			setup:     func() { listItemsOpts.Untriaged = true; listItemsOpts.TimeSensitive = true },
			wantError: true,
		},
		{
			name:      "untriaged and done are mutually exclusive",
			setup:     func() { listItemsOpts.Untriaged = true; listItemsOpts.Done = true },
			wantError: true,
		},
		{
			name:      "time-sensitive and done are mutually exclusive",
			setup:     func() { listItemsOpts.TimeSensitive = true; listItemsOpts.Done = true },
			wantError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Cleanup(resetListItemsOpts)
			test.setup()

			got, err := resolveItemView()
			if test.wantError {
				if err == nil {
					t.Errorf("expected an error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error but got %v", err)
			}
			if got != test.expected {
				t.Errorf("expected %v but got %v", test.expected, got)
			}
		})
	}
}

func TestViewHeader(t *testing.T) {
	tests := []struct {
		name          string
		view          proto.ItemView
		expectedHead  string
		expectNumbered bool
	}{
		{"triaged default", proto.ItemView_ITEM_VIEW_TRIAGED, "Items:", true},
		{"untriaged", proto.ItemView_ITEM_VIEW_UNTRIAGED, "Untriaged:", true},
		{"time-sensitive", proto.ItemView_ITEM_VIEW_TIME_SENSITIVE, "Time-sensitive:", true},
		{"done", proto.ItemView_ITEM_VIEW_DONE, "Done:", false},
		{"unspecified falls back to items", proto.ItemView_ITEM_VIEW_UNSPECIFIED, "Items:", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			head, numbered := viewHeader(test.view)
			if head != test.expectedHead {
				t.Errorf("expected header %q but got %q", test.expectedHead, head)
			}
			if numbered != test.expectNumbered {
				t.Errorf("expected numbered %v but got %v", test.expectNumbered, numbered)
			}
		})
	}
}
