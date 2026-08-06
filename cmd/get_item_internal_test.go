package cmd

import (
	"bytes"
	"strings"
	"testing"
	"time"

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

func TestFormatPriority(t *testing.T) {
	tests := []struct {
		name     string
		priority float64
		expected string
	}{
		{"whole number", 2.0, "2"},
		{"one decimal", 2.5, "2.5"},
		{"two decimals", 2.25, "2.25"},
		{"zero", 0.0, "0"},
		{"max uint32", 4294967295.0, "4294967295"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := formatPriority(test.priority); got != test.expected {
				t.Errorf("expected %q but got %q", test.expected, got)
			}
		})
	}
}

func TestWriteItemDetailRendersDueDate(t *testing.T) {
	var buf bytes.Buffer
	due := timestamppb.New(time.Date(2026, time.August, 15, 0, 0, 0, 0, time.UTC))
	if err := writeItemDetail(&buf, &proto.Item{Id: 1, Title: "due soon", DueDate: due, Priority: floatPtr(1.0)}); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Due:") {
		t.Errorf("expected a Due: row in %q", out)
	}
	if !strings.Contains(out, "2026-08-15") {
		t.Errorf("expected the formatted due date in %q", out)
	}
}

func TestWriteCommentDetailEmptyCreatedAndAuthor(t *testing.T) {
	var buf bytes.Buffer
	item := &proto.Item{
		Id: 1,
		Title: "discussed",
		Comments: []*proto.Comment{
			{Id: 1, Body: "no timestamp", Author: "alex"},
			{Id: 2, Body: "anonymous", Author: "", CreatedAt: timestamppb.Now()},
		},
	}
	if err := writeCommentDetail(&buf, item); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Comments (2):") {
		t.Errorf("expected a Comments header in %q", out)
	}
	if !strings.Contains(out, "no timestamp") || !strings.Contains(out, "alex") {
		t.Errorf("expected the first comment body and author in %q", out)
	}
	if !strings.Contains(out, "anonymous") {
		t.Errorf("expected the second comment body in %q", out)
	}
}

func TestWriteBlockerDetailEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := writeBlockerDetail(&buf, &proto.Item{Id: 1, Title: "bare"}); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output for an item with no blockers but got %q", buf.String())
	}
}

func TestWriteCommentDetailEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := writeCommentDetail(&buf, &proto.Item{Id: 1, Title: "bare"}); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output for an item with no comments but got %q", buf.String())
	}
}

// failingWriter is an io.Writer that always returns an error, used to
// exercise the error-return branches of the output helpers.
type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errWriteFailed
}

var errWriteFailed = bytesErr("write failed")

type bytesErr string

func (e bytesErr) Error() string { return string(e) }

func TestWriteItemDetailWriteError(t *testing.T) {
	if err := writeItemDetail(failingWriter{}, &proto.Item{Id: 1, Title: "a"}); err == nil {
		t.Errorf("expected a write error but got none")
	}
}

func TestWriteBlockerDetailWriteError(t *testing.T) {
	item := &proto.Item{
		Id:       1,
		Title:    "a",
		Blockers: []*proto.Blocker{{Id: 1, Description: "waiting"}},
	}
	if err := writeBlockerDetail(failingWriter{}, item); err == nil {
		t.Errorf("expected a write error but got none")
	}
}

func TestWriteCommentDetailWriteError(t *testing.T) {
	item := &proto.Item{
		Id:       1,
		Title:    "a",
		Comments: []*proto.Comment{{Id: 1, Body: "note", Author: "alex", CreatedAt: timestamppb.Now()}},
	}
	if err := writeCommentDetail(failingWriter{}, item); err == nil {
		t.Errorf("expected a write error but got none")
	}
}

func floatPtr(v float64) *float64 { return &v }