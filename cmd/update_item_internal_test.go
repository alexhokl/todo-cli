package cmd

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestUpdateItemFlagValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{"add only", []string{"7", "--add-label", "urgent"}, false},
		{"remove only", []string{"7", "--remove-label", "later"}, false},
		{"add and remove", []string{"7", "--add-label", "urgent", "--remove-label", "later"}, false},
		{"repeated add", []string{"7", "--add-label", "urgent", "--add-label", "work"}, false},
		{"no label flags", []string{"7"}, true},
		{"no arguments", []string{"--add-label", "urgent"}, true},
		{"too many arguments", []string{"7", "8", "--add-label", "urgent"}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// The flag rules are validated before RunE is reached, so a stub
			// run function keeps the assertion off the network.
			cmd := newUpdateItemCmd()
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

func TestJoinLabelNames(t *testing.T) {
	tests := []struct {
		name     string
		labels   []*proto.Label
		expected string
	}{
		{"none", nil, "-"},
		{"one", []*proto.Label{{Name: "work"}}, "work"},
		{"several", []*proto.Label{{Name: "work"}, {Name: "urgent"}}, "work,urgent"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := joinLabelNames(&proto.Item{Labels: test.labels})
			if result != test.expected {
				t.Errorf("expected %q but got %q", test.expected, result)
			}
		})
	}
}

func TestWriteItemTableIncludesLabels(t *testing.T) {
	var buffer strings.Builder
	items := []*proto.Item{
		{Id: 1, Title: "a", Labels: []*proto.Label{{Name: "work"}}},
		{Id: 2, Title: "b"},
	}

	if err := writeItemTable(&buffer, items, true); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	output := buffer.String()
	if !strings.Contains(output, "LABELS") {
		t.Errorf("expected a LABELS column but got:\n%s", output)
	}
	if !strings.Contains(output, "work") {
		t.Errorf("expected the label to be rendered but got:\n%s", output)
	}
}

func TestWriteItemLineIncludesLabels(t *testing.T) {
	var buffer strings.Builder
	item := &proto.Item{Id: 7, Title: "a", Labels: []*proto.Label{{Name: "work"}, {Name: "urgent"}}}

	if err := writeItemLine(&buffer, item); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if !strings.Contains(buffer.String(), "labels work,urgent") {
		t.Errorf("expected the labels to be rendered but got %q", buffer.String())
	}
}

func TestWriteItemTableEmpty(t *testing.T) {
	var buffer strings.Builder
	if err := writeItemTable(&buffer, nil, true); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if buffer.String() != "  (none)\n" {
		t.Errorf("expected the none placeholder but got %q", buffer.String())
	}
}

func TestWriteItemTableRendersDueDateAndList(t *testing.T) {
	var buffer strings.Builder
	due := timestamppb.New(time.Date(2026, time.August, 15, 0, 0, 0, 0, time.UTC))
	listID := uint32(3)
	items := []*proto.Item{
		{Id: 1, Title: "a", DueDate: due, ListId: &listID},
	}

	if err := writeItemTable(&buffer, items, true); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	output := buffer.String()
	// The due date is formatted in the local time zone, so only assert the
	// date portion is present rather than a full timestamp.
	if !strings.Contains(output, "2026-08-15") {
		t.Errorf("expected the due date to be rendered but got:\n%s", output)
	}
	if !strings.Contains(output, " 3 ") {
		t.Errorf("expected the list id 3 to be rendered but got:\n%s", output)
	}
}

func TestWriteItemTableUnnumberedShowsDash(t *testing.T) {
	var buffer strings.Builder
	items := []*proto.Item{{Id: 1, Title: "a"}}

	if err := writeItemTable(&buffer, items, false); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	// The first data row should lead with a dash rather than an ordinal.
	lines := strings.Split(strings.TrimRight(buffer.String(), "\n"), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least a header and one row but got %d lines", len(lines))
	}
	if !strings.HasPrefix(lines[1], "-") {
		t.Errorf("expected the unnumbered row to start with a dash but got %q", lines[1])
	}
	// A numbered row would lead with "1", so confirm it does not.
	if strings.HasPrefix(lines[1], "1") {
		t.Errorf("expected the unnumbered row not to start with an ordinal but got %q", lines[1])
	}
}

func TestWriteItemLineWithDoneAndList(t *testing.T) {
	var buffer strings.Builder
	listID := uint32(2)
	item := &proto.Item{Id: 7, Title: "a", Done: true, ListId: &listID}

	if err := writeItemLine(&buffer, item); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	output := buffer.String()
	if !strings.Contains(output, "list 2") {
		t.Errorf("expected the list id to be rendered but got %q", output)
	}
	if !strings.Contains(output, "done") {
		t.Errorf("expected the done marker to be rendered but got %q", output)
	}
}
