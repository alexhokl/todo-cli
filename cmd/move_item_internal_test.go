package cmd

import (
	"io"
	"math"
	"testing"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

func TestParseID(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    uint32
		expectError bool
	}{
		{"valid", "7", 7, false},
		{"largest valid", "4294967295", 4294967295, false},
		{"zero", "0", 0, true},
		{"negative", "-1", 0, true},
		{"out of range", "4294967296", 0, true},
		{"not a number", "seven", 0, true},
		{"empty", "", 0, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id, err := parseID(test.input, "item")
			if test.expectError {
				if err == nil {
					t.Errorf("expected an error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("expected no error but got %v", err)
			}
			if id != test.expected {
				t.Errorf("expected %d but got %d", test.expected, id)
			}
		})
	}
}

func TestItemCommandsRequireService(t *testing.T) {
	commands := map[string]bool{
		"update priority":   requiresService(moveItemCmd),
		"list items":        requiresService(listItemsCmd),
		"get item":          requiresService(getItemCmd),
		"create item":       requiresService(createItemCmd),
		"update done":       requiresService(completeItemCmd),
		"update item":       requiresService(updateItemCmd),
		"update todo effort": requiresService(setItemEffortCmd),
		"list labels":       requiresService(listLabelsCmd),
		"create label":      requiresService(createLabelCmd),
		"update label":      requiresService(renameLabelCmd),
		"delete label":      requiresService(deleteLabelCmd),
		"list efforts":      requiresService(listEffortsCmd),
		"create effort":     requiresService(createEffortCmd),
		"update effort":     requiresService(renameEffortCmd),
		"delete effort":     requiresService(deleteEffortCmd),
		"list blockers":     requiresService(listBlockersCmd),
		"create blocker":    requiresService(createBlockerCmd),
		"update blocker":    requiresService(updateBlockerCmd),
		"delete blocker":    requiresService(deleteBlockerCmd),
	}

	for name, required := range commands {
		t.Run(name, func(t *testing.T) {
			if !required {
				t.Errorf("expected the %q command to require a service", name)
			}
		})
	}
}
func TestToUint32Slice(t *testing.T) {
	tests := []struct {
		name    string
		ids     []uint
		expected []uint32
		wantError bool
	}{
		{"empty", []uint{}, []uint32{}, false},
		{"single in range", []uint{7}, []uint32{7}, false},
		{"several in range", []uint{1, 2, 3}, []uint32{1, 2, 3}, false},
		{"max uint32", []uint{math.MaxUint32}, []uint32{math.MaxUint32}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := toUint32Slice(test.ids)
			if test.wantError {
				if err == nil {
					t.Fatalf("expected an error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error but got %v", err)
			}
			if len(result) != len(test.expected) {
				t.Fatalf("expected %v but got %v", test.expected, result)
			}
			for i, v := range result {
				if v != test.expected[i] {
					t.Errorf("at index %d expected %d but got %d", i, test.expected[i], v)
				}
			}
		})
	}

	t.Run("out of range", func(t *testing.T) {
		over, ok := maxUint32Plus1()
		if !ok {
			t.Skip("uint is 32-bit on this platform; overflow case is unreachable")
		}
		if _, err := toUint32Slice([]uint{over}); err == nil {
			t.Errorf("expected an error but got none")
		}
	})

	t.Run("mixed valid and out of range", func(t *testing.T) {
		over, ok := maxUint32Plus1()
		if !ok {
			t.Skip("uint is 32-bit on this platform; overflow case is unreachable")
		}
		if _, err := toUint32Slice([]uint{1, over}); err == nil {
			t.Errorf("expected an error but got none")
		}
	})
}

// maxUint32Plus1 returns a value greater than math.MaxUint32, or 0 and false
// on 32-bit platforms where uint cannot hold such a value. Shared across the
// cmd package tests so the overflow path is exercised only where it exists.
func maxUint32Plus1() (uint, bool) {
	if ^uint(0)>>32 == 0 {
		return 0, false
	}
	return uint(math.MaxUint32) + 1, true
}

func TestBuildMoveItemRequestTopBottom(t *testing.T) {
	t.Run("top anchor", func(t *testing.T) {
		req := buildMoveItemRequest(7, moveItemOptions{Top: true}, false)
		if req.GetId() != 7 {
			t.Errorf("expected id 7 but got %d", req.GetId())
		}
		if top, ok := req.GetAnchor().(*proto.MoveItemRequest_Top); !ok || !top.Top {
			t.Errorf("expected a Top anchor set to true but got %T %v", req.GetAnchor(), req.GetAnchor())
		}
	})

	t.Run("bottom anchor", func(t *testing.T) {
		req := buildMoveItemRequest(7, moveItemOptions{Bottom: true}, false)
		if req.GetId() != 7 {
			t.Errorf("expected id 7 but got %d", req.GetId())
		}
		if bottom, ok := req.GetAnchor().(*proto.MoveItemRequest_Bottom); !ok || !bottom.Bottom {
			t.Errorf("expected a Bottom anchor set to true but got %T %v", req.GetAnchor(), req.GetAnchor())
		}
	})
}

func TestMoveItemFlagValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{"before only", []string{"7", "--before", "3"}, false},
		{"after only", []string{"7", "--after", "3"}, false},
		{"top only", []string{"7", "--top"}, false},
		{"bottom only", []string{"7", "--bottom"}, false},
		{"before with list", []string{"7", "--before", "3", "--list", "2"}, false},
		{"after with clear-list", []string{"7", "--after", "3", "--clear-list"}, false},
		{"top with list", []string{"7", "--top", "--list", "2"}, false},
		{"bottom with clear-list", []string{"7", "--bottom", "--clear-list"}, false},
		{"no anchor", []string{"7"}, true},
		{"both anchors", []string{"7", "--before", "3", "--after", "4"}, true},
		{"before with top", []string{"7", "--before", "3", "--top"}, true},
		{"after with bottom", []string{"7", "--after", "3", "--bottom"}, true},
		{"top with bottom", []string{"7", "--top", "--bottom"}, true},
		{"before with bottom", []string{"7", "--before", "3", "--bottom"}, true},
		{"list and clear-list", []string{"7", "--before", "3", "--list", "2", "--clear-list"}, true},
		{"no arguments", []string{"--before", "3"}, true},
		{"too many arguments", []string{"7", "8", "--before", "3"}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// The flag rules are validated before RunE is reached, so a stub
			// run function keeps the assertion off the network.
			cmd := newMoveItemCmd()
			cmd.RunE = func(_ *cobra.Command, _ []string) error { return nil }
			cmd.SetArgs(test.args)
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)

			err := cmd.Execute()
			if test.expectError && err == nil {
				t.Errorf("expected an error but got none")
			}
			if !test.expectError && err != nil {
				t.Errorf("expected no error but got %v", err)
			}
		})
	}
}
