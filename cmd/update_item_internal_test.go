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
		{"set due date", []string{"7", "--due-date", "2026-08-15"}, false},
		{"clear due date", []string{"7", "--clear-due-date"}, false},
		{"both due date flags", []string{"7", "--due-date", "2026-08-15", "--clear-due-date"}, true},
		{"repeated add", []string{"7", "--add-label", "urgent", "--add-label", "work"}, false},
		{"title only", []string{"7", "--title", "renamed"}, false},
		{"description only", []string{"7", "--description", "new details"}, false},
		{"title and description", []string{"7", "--title", "renamed", "--description", "new details"}, false},
		{"clear description", []string{"7", "--description", ""}, false},
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

func TestParseDueDate(t *testing.T) {
	valid, err := parseDueDate("2026-08-15")
	if err != nil {
		t.Fatalf("expected valid date but got %v", err)
	}
	if valid.Format(time.DateOnly) != "2026-08-15" {
		t.Errorf("expected 2026-08-15 but got %s", valid.Format(time.DateOnly))
	}

	if _, err := parseDueDate("15-08-2026"); err == nil {
		t.Errorf("expected invalid date error")
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

	if err := writeItemTable(&buffer, items); err != nil {
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
	if err := writeItemTable(&buffer, nil); err != nil {
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

	if err := writeItemTable(&buffer, items); err != nil {
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

func TestBuildUpdateItemLinksRequest(t *testing.T) {
	t.Run("empty add and remove", func(t *testing.T) {
		req, err := buildUpdateItemLinksRequest(7, updateItemOptions{})
		if err != nil {
			t.Fatalf("expected no error but got %v", err)
		}
		if req.GetId() != 7 {
			t.Errorf("expected id 7 but got %d", req.GetId())
		}
		if len(req.GetAdd()) != 0 {
			t.Errorf("expected no add ids but got %v", req.GetAdd())
		}
		if len(req.GetRemove()) != 0 {
			t.Errorf("expected no remove ids but got %v", req.GetRemove())
		}
	})

	t.Run("add only", func(t *testing.T) {
		req, err := buildUpdateItemLinksRequest(7, updateItemOptions{AddLinks: []uint{3, 5}})
		if err != nil {
			t.Fatalf("expected no error but got %v", err)
		}
		if len(req.GetAdd()) != 2 || req.GetAdd()[0] != 3 || req.GetAdd()[1] != 5 {
			t.Errorf("expected add [3 5] but got %v", req.GetAdd())
		}
		if len(req.GetRemove()) != 0 {
			t.Errorf("expected no remove ids but got %v", req.GetRemove())
		}
	})

	t.Run("remove only", func(t *testing.T) {
		req, err := buildUpdateItemLinksRequest(7, updateItemOptions{RemoveLinks: []uint{2}})
		if err != nil {
			t.Fatalf("expected no error but got %v", err)
		}
		if len(req.GetRemove()) != 1 || req.GetRemove()[0] != 2 {
			t.Errorf("expected remove [2] but got %v", req.GetRemove())
		}
		if len(req.GetAdd()) != 0 {
			t.Errorf("expected no add ids but got %v", req.GetAdd())
		}
	})

	t.Run("add and remove", func(t *testing.T) {
		req, err := buildUpdateItemLinksRequest(7, updateItemOptions{AddLinks: []uint{3}, RemoveLinks: []uint{9}})
		if err != nil {
			t.Fatalf("expected no error but got %v", err)
		}
		if len(req.GetAdd()) != 1 || req.GetAdd()[0] != 3 {
			t.Errorf("expected add [3] but got %v", req.GetAdd())
		}
		if len(req.GetRemove()) != 1 || req.GetRemove()[0] != 9 {
			t.Errorf("expected remove [9] but got %v", req.GetRemove())
		}
	})

	t.Run("out of range add id", func(t *testing.T) {
		over, ok := maxUint32Plus1()
		if !ok {
			t.Skip("uint is 32-bit on this platform; overflow case is unreachable")
		}
		_, err := buildUpdateItemLinksRequest(7, updateItemOptions{AddLinks: []uint{over}})
		if err == nil {
			t.Fatalf("expected an error but got none")
		}
	})

	t.Run("out of range remove id", func(t *testing.T) {
		over, ok := maxUint32Plus1()
		if !ok {
			t.Skip("uint is 32-bit on this platform; overflow case is unreachable")
		}
		_, err := buildUpdateItemLinksRequest(7, updateItemOptions{RemoveLinks: []uint{over}})
		if err == nil {
			t.Fatalf("expected an error but got none")
		}
	})
}

func TestJoinBlockerDescriptions(t *testing.T) {
	tests := []struct {
		name     string
		blockers []*proto.Blocker
		expected string
	}{
		{"none", nil, "-"},
		{"one", []*proto.Blocker{{Description: "waiting"}}, "waiting"},
		{"several", []*proto.Blocker{{Description: "waiting"}, {Description: "needs review"}}, "waiting; needs review"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := joinBlockerDescriptions(&proto.Item{Blockers: test.blockers})
			if result != test.expected {
				t.Errorf("expected %q but got %q", test.expected, result)
			}
		})
	}
}

func TestJoinLinkedItemIDs(t *testing.T) {
	tests := []struct {
		name     string
		linked   []*proto.Item
		expected string
	}{
		{"none", nil, "-"},
		{"one", []*proto.Item{{Id: 3}}, "3"},
		{"several", []*proto.Item{{Id: 3}, {Id: 5}}, "3,5"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := joinLinkedItemIDs(&proto.Item{LinkedItems: test.linked})
			if result != test.expected {
				t.Errorf("expected %q but got %q", test.expected, result)
			}
		})
	}
}

func TestCommentCount(t *testing.T) {
	tests := []struct {
		name     string
		comments []*proto.Comment
		expected string
	}{
		{"none", nil, "-"},
		{"one", []*proto.Comment{{Id: 1}}, "1"},
		{"several", []*proto.Comment{{Id: 1}, {Id: 2}}, "2"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := commentCount(&proto.Item{Comments: test.comments})
			if result != test.expected {
				t.Errorf("expected %q but got %q", test.expected, result)
			}
		})
	}
}

func TestEffortName(t *testing.T) {
	t.Run("nil effort", func(t *testing.T) {
		if got := effortName(&proto.Item{Effort: nil}); got != "-" {
			t.Errorf("expected %q but got %q", "-", got)
		}
	})

	t.Run("named effort", func(t *testing.T) {
		item := &proto.Item{Effort: &proto.Effort{Id: 1, Name: "high"}}
		if got := effortName(item); got != "high" {
			t.Errorf("expected %q but got %q", "high", got)
		}
	})
}

func TestWriteItemLineFull(t *testing.T) {
	var buffer strings.Builder
	listID := uint32(2)
	item := &proto.Item{
		Id:          7,
		Title:       "ship it",
		Done:        true,
		ListId:      &listID,
		Effort:      &proto.Effort{Id: 1, Name: "high"},
		Labels:      []*proto.Label{{Name: "urgent"}},
		Blockers:    []*proto.Blocker{{Id: 1, Description: "waiting"}},
		Comments:    []*proto.Comment{{Id: 1}, {Id: 2}},
		LinkedItems: []*proto.Item{{Id: 3}, {Id: 5}},
	}

	if err := writeItemLine(&buffer, item); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	output := buffer.String()
	for _, want := range []string{
		"ship it (",
		"id 7",
		"labels urgent",
		"effort high",
		"blockers 1",
		"comments 2",
		"linked 3,5",
		"list 2",
		"done",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("expected %q in the output but it is missing in %q", want, output)
		}
	}
}

func TestWriteItemTableFullRow(t *testing.T) {
	var buffer strings.Builder
	listID := uint32(2)
	due := timestamppb.New(time.Date(2026, time.August, 15, 0, 0, 0, 0, time.UTC))
	items := []*proto.Item{
		{
			Id:          1,
			Title:       "ship it",
			DueDate:     due,
			ListId:      &listID,
			Effort:      &proto.Effort{Id: 1, Name: "high"},
			Labels:      []*proto.Label{{Name: "urgent"}},
			Blockers:    []*proto.Blocker{{Id: 1, Description: "waiting"}},
			Comments:    []*proto.Comment{{Id: 1}, {Id: 2}},
			LinkedItems: []*proto.Item{{Id: 3}, {Id: 5}},
		},
	}

	if err := writeItemTable(&buffer, items); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	output := buffer.String()
	// The tabwriter expands tabs to spaces, so assert the header labels and
	// the row's content are present rather than an exact tab-delimited line.
	for _, want := range []string{
		"ID",
		"TITLE",
		"LABELS",
		"EFFORT",
		"BLOCKERS",
		"COMMENTS",
		"LINKED",
		"LIST",
		"DUE",
		"ship it",
		"urgent",
		"high",
		"waiting",
		"3,5",
		"2026-08-15",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("expected %q in the output but it is missing in %q", want, output)
		}
	}
}

func TestWriteItemTableWriteError(t *testing.T) {
	items := []*proto.Item{{Id: 1, Title: "a"}}
	if err := writeItemTable(failingWriter{}, items); err == nil {
		t.Errorf("expected a write error but got none")
	}
}

func TestWriteItemLineWriteError(t *testing.T) {
	if err := writeItemLine(failingWriter{}, &proto.Item{Id: 1, Title: "a"}); err == nil {
		t.Errorf("expected a write error but got none")
	}
}
