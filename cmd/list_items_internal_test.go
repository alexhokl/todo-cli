package cmd

import (
	"strings"
	"testing"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
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

// TestListItemsSearchFlagBinding asserts the --search flag is registered on
// the command, parses into listItemsOpts.Search, and defaults to the empty
// no-op string. The gRPC request construction itself is not exercised here (it
// would require a dial stub); instead the flag binding is verified directly.
func TestListItemsSearchFlagBinding(t *testing.T) {
	t.Cleanup(resetListItemsOpts)

	flag := listItemsCmd.Flags().Lookup("search")
	if flag == nil {
		t.Fatalf("expected the --search flag to be registered")
	}
	if flag.DefValue != "" {
		t.Errorf("expected the --search default to be empty but got %q", flag.DefValue)
	}

	// Parse a command line carrying the flag and confirm it lands on the
	// options struct. A dummy RunE prevents execution of the real dial path.
	cmd := &cobra.Command{
		Use: "items",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	cmd.Flags().AddFlagSet(listItemsCmd.Flags())

	if err := cmd.ParseFlags([]string{"--search", "milk"}); err != nil {
		t.Fatalf("failed to parse the --search flag: %v", err)
	}
	// AddFlagSet shares the same flag pointers, so the parsed value is
	// reflected on listItemsOpts.Search.
	if listItemsOpts.Search != "milk" {
		t.Errorf("expected listItemsOpts.Search to be %q but got %q", "milk", listItemsOpts.Search)
	}
}

// TestListItemsSearchFlagInHelp asserts the long help documents --search so
// users can discover it.
func TestListItemsSearchFlagInHelp(t *testing.T) {
	if !strings.Contains(listItemsCmd.Long, "--search") {
		t.Errorf("expected the long help to mention --search but got: %s", listItemsCmd.Long)
	}
}
