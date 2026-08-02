package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/alexhokl/todo-cli/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestItemDetailHeader(t *testing.T) {
	tests := []struct {
		name     string
		item     *proto.Item
		expected string
	}{
		{"triaged active", &proto.Item{Id: 7, Title: "ship it", Priority: floatPtr(2.0)}, "Item 7: ship it"},
		{"untriaged", &proto.Item{Id: 7, Title: "ship it"}, "Item 7 (untriaged): ship it"},
		{"done", &proto.Item{Id: 7, Title: "ship it", Done: true}, "Item 7 (done): ship it"},
		{"done ignores priority", &proto.Item{Id: 7, Title: "ship it", Done: true, Priority: floatPtr(2.0)}, "Item 7 (done): ship it"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := itemDetailHeader(test.item)
			if got != test.expected {
				t.Errorf("expected %q but got %q", test.expected, got)
			}
		})
	}
}

func TestWriteItemDetailOmitsEmptyOptionalFields(t *testing.T) {
	var buf bytes.Buffer
	if err := writeItemDetail(&buf, &proto.Item{Id: 1, Title: "bare", Priority: floatPtr(1.0)}); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	out := buf.String()
	if !strings.HasPrefix(out, "Item 1: bare\n\n") {
		t.Errorf("expected the header line first but got %q", out)
	}
	for _, field := range []string{"Description", "Due", "List", "Effort", "Labels", "Linked", "Blockers", "Comments"} {
		if strings.Contains(out, field) {
			t.Errorf("expected %q to be omitted for a bare item but it appears in %q", field, out)
		}
	}
	if !strings.Contains(out, "Priority:") {
		t.Errorf("expected a Priority row for an active triaged item but it is missing in %q", out)
	}
}

func TestWriteItemDetailUntriagedShowsPriorityDash(t *testing.T) {
	var buf bytes.Buffer
	if err := writeItemDetail(&buf, &proto.Item{Id: 1, Title: "new", Priority: nil}); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Priority:  -") {
		t.Errorf("expected \"Priority:  -\" for an untriaged item but got %q", out)
	}
}

func TestWriteItemDetailDoneOmitsPriority(t *testing.T) {
	var buf bytes.Buffer
	if err := writeItemDetail(&buf, &proto.Item{Id: 1, Title: "done", Done: true, Priority: nil}); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "Priority") {
		t.Errorf("expected no Priority row for a done item but it appears in %q", out)
	}
}

func TestWriteItemDetailRendersAllFields(t *testing.T) {
	listID := uint32(3)
	item := &proto.Item{
		Id:          7,
		Title:       "ship release notes",
		Description: "draft the 0.2 changelog",
		Priority:    floatPtr(2.0),
		ListId:      &listID,
		Effort:      &proto.Effort{Id: 1, Name: "high"},
		Labels:      []*proto.Label{{Id: 1, Name: "urgent"}, {Id: 2, Name: "work"}},
		LinkedItems: []*proto.Item{{Id: 3, Title: "release checklist"}, {Id: 5, Title: "notify customers"}},
		Blockers: []*proto.Blocker{
			{Id: 1, Description: "waiting on legal review"},
			{Id: 2, Description: "needs design sign-off"},
		},
		Comments: []*proto.Comment{
			{Id: 1, Body: "drafted a reply", Author: "alex", CreatedAt: timestamppb.Now()},
			{Id: 2, Body: "approved", Author: "sara", CreatedAt: timestamppb.Now()},
		},
	}

	var buf bytes.Buffer
	if err := writeItemDetail(&buf, item); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	out := buf.String()

	for _, want := range []string{
		"Item 7: ship release notes",
		"Description:",
		"draft the 0.2 changelog",
		"List:",
		"3",
		"Priority:",
		"Effort:",
		"high",
		"Labels:",
		"urgent,work",
		"Linked:",
		"3 - release checklist, 5 - notify customers",
		"Blockers (2):",
		"waiting on legal review",
		"needs design sign-off",
		"Comments (2):",
		"drafted a reply",
		"approved",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in the output but it is missing in %q", want, out)
		}
	}
}

func TestJoinLinkedItemSummaries(t *testing.T) {
	item := &proto.Item{
		LinkedItems: []*proto.Item{{Id: 3, Title: "a"}, {Id: 5, Title: "b"}},
	}
	got := joinLinkedItemSummaries(item)
	want := "3 - a, 5 - b"
	if got != want {
		t.Errorf("expected %q but got %q", want, got)
	}

	empty := joinLinkedItemSummaries(&proto.Item{})
	if empty != "" {
		t.Errorf("expected an empty string for no linked items but got %q", empty)
	}
}

func floatPtr(v float64) *float64 { return &v }