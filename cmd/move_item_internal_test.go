package cmd

import (
	"io"
	"testing"

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
		"update priority": requiresService(moveItemCmd),
		"list items":      requiresService(listItemsCmd),
		"create item":     requiresService(createItemCmd),
		"update done":     requiresService(completeItemCmd),
		"update item":     requiresService(updateItemCmd),
		"list labels":     requiresService(listLabelsCmd),
		"create label":    requiresService(createLabelCmd),
		"update label":    requiresService(renameLabelCmd),
		"delete label":    requiresService(deleteLabelCmd),
	}

	for name, required := range commands {
		t.Run(name, func(t *testing.T) {
			if !required {
				t.Errorf("expected the %q command to require a service", name)
			}
		})
	}
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
