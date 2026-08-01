package internal

import (
	"context"
	"testing"

	"github.com/alexhokl/todo-cli/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// itemLabelNames returns the label names carried by an item.
func itemLabelNames(item *proto.Item) []string {
	names := make([]string, 0, len(item.GetLabels()))
	for _, label := range item.GetLabels() {
		names = append(names, label.GetName())
	}

	return names
}

func containsAll(names []string, expected ...string) bool {
	index := make(map[string]struct{}, len(names))
	for _, name := range names {
		index[name] = struct{}{}
	}
	if len(index) != len(expected) {
		return false
	}
	for _, name := range expected {
		if _, ok := index[name]; !ok {
			return false
		}
	}

	return true
}

func TestCreateLabelNormalisesAndRejectsDuplicates(t *testing.T) {
	server := setupItemServer(t)

	label, err := server.CreateLabel(authenticatedContext(), &proto.CreateLabelRequest{Name: "  Work  "})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if label.GetName() != "work" {
		t.Errorf("expected %q but got %q", "work", label.GetName())
	}

	_, err = server.CreateLabel(authenticatedContext(), &proto.CreateLabelRequest{Name: "WORK"})
	if got := status.Code(err); got != codes.AlreadyExists {
		t.Errorf("expected %v but got %v (%v)", codes.AlreadyExists, got, err)
	}

	_, err = server.CreateLabel(authenticatedContext(), &proto.CreateLabelRequest{Name: "   "})
	if got := status.Code(err); got != codes.InvalidArgument {
		t.Errorf("expected %v but got %v (%v)", codes.InvalidArgument, got, err)
	}
}

func TestListLabels(t *testing.T) {
	server := setupItemServer(t)
	for _, name := range []string{"work", "admin"} {
		if _, err := server.CreateLabel(authenticatedContext(), &proto.CreateLabelRequest{Name: name}); err != nil {
			t.Fatalf("failed to create the label %q: %v", name, err)
		}
	}

	response, err := server.ListLabels(authenticatedContext(), &proto.ListLabelsRequest{})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if len(response.GetLabels()) != 2 {
		t.Fatalf("expected 2 labels but got %d", len(response.GetLabels()))
	}
	if response.GetLabels()[0].GetName() != "admin" {
		t.Errorf("expected the labels to be ordered by name but got %v", response.GetLabels())
	}
}

func TestRenameLabelErrorCodes(t *testing.T) {
	server := setupItemServer(t)
	work, err := server.CreateLabel(authenticatedContext(), &proto.CreateLabelRequest{Name: "work"})
	if err != nil {
		t.Fatalf("failed to create the label: %v", err)
	}
	if _, err := server.CreateLabel(authenticatedContext(), &proto.CreateLabelRequest{Name: "home"}); err != nil {
		t.Fatalf("failed to create the label: %v", err)
	}

	tests := []struct {
		name     string
		id       uint32
		newName  string
		expected codes.Code
	}{
		{"missing id", 0, "office", codes.InvalidArgument},
		{"unknown id", 404, "office", codes.NotFound},
		{"empty name", work.GetId(), "  ", codes.InvalidArgument},
		{"name already taken", work.GetId(), "home", codes.AlreadyExists},
		{"valid", work.GetId(), "office", codes.OK},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := &proto.RenameLabelRequest{Id: test.id, Name: test.newName}
			_, err := server.RenameLabel(authenticatedContext(), req)
			if got := status.Code(err); got != test.expected {
				t.Errorf("expected %v but got %v (%v)", test.expected, got, err)
			}
		})
	}
}

func TestDeleteLabelRefusesWhileInUse(t *testing.T) {
	server := setupItemServer(t)
	ids := createItems(t, server, "a")
	label, err := server.CreateLabel(authenticatedContext(), &proto.CreateLabelRequest{Name: "work"})
	if err != nil {
		t.Fatalf("failed to create the label: %v", err)
	}

	tagged, err := server.UpdateItemLabels(authenticatedContext(), &proto.UpdateItemLabelsRequest{
		Id:  ids["a"],
		Add: []string{"work"},
	})
	if err != nil {
		t.Fatalf("failed to tag the item: %v", err)
	}
	if !containsAll(itemLabelNames(tagged), "work") {
		t.Errorf("expected [work] but got %v", itemLabelNames(tagged))
	}

	_, err = server.DeleteLabel(authenticatedContext(), &proto.DeleteLabelRequest{Id: label.GetId()})
	if got := status.Code(err); got != codes.FailedPrecondition {
		t.Errorf("expected %v but got %v (%v)", codes.FailedPrecondition, got, err)
	}

	if _, err := server.UpdateItemLabels(authenticatedContext(), &proto.UpdateItemLabelsRequest{
		Id:     ids["a"],
		Remove: []string{"work"},
	}); err != nil {
		t.Fatalf("failed to untag the item: %v", err)
	}

	if _, err := server.DeleteLabel(authenticatedContext(), &proto.DeleteLabelRequest{Id: label.GetId()}); err != nil {
		t.Errorf("expected no error but got %v", err)
	}
}

func TestDeleteLabelErrorCodes(t *testing.T) {
	tests := []struct {
		name     string
		id       uint32
		expected codes.Code
	}{
		{"missing id", 0, codes.InvalidArgument},
		{"unknown id", 404, codes.NotFound},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := setupItemServer(t)
			_, err := server.DeleteLabel(authenticatedContext(), &proto.DeleteLabelRequest{Id: test.id})
			if got := status.Code(err); got != test.expected {
				t.Errorf("expected %v but got %v (%v)", test.expected, got, err)
			}
		})
	}
}

func TestUpdateItemLabelsErrorCodes(t *testing.T) {
	tests := []struct {
		name     string
		request  func(ids map[string]uint32) *proto.UpdateItemLabelsRequest
		expected codes.Code
	}{
		{
			name: "missing id",
			request: func(_ map[string]uint32) *proto.UpdateItemLabelsRequest {
				return &proto.UpdateItemLabelsRequest{Add: []string{"work"}}
			},
			expected: codes.InvalidArgument,
		},
		{
			name: "neither add nor remove",
			request: func(ids map[string]uint32) *proto.UpdateItemLabelsRequest {
				return &proto.UpdateItemLabelsRequest{Id: ids["a"]}
			},
			expected: codes.InvalidArgument,
		},
		{
			name: "unknown item",
			request: func(_ map[string]uint32) *proto.UpdateItemLabelsRequest {
				return &proto.UpdateItemLabelsRequest{Id: 404, Add: []string{"work"}}
			},
			expected: codes.NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := setupItemServer(t)
			ids := createItems(t, server, "a")
			_, err := server.UpdateItemLabels(authenticatedContext(), test.request(ids))
			if got := status.Code(err); got != test.expected {
				t.Errorf("expected %v but got %v (%v)", test.expected, got, err)
			}
		})
	}
}

func TestCreateItemWithLabels(t *testing.T) {
	server := setupItemServer(t)

	item, err := server.CreateItem(authenticatedContext(), &proto.CreateItemRequest{
		Title:  "ship it",
		Labels: []string{"Urgent", "  urgent ", "work"},
	})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if !containsAll(itemLabelNames(item), "urgent", "work") {
		t.Errorf("expected [urgent work] but got %v", itemLabelNames(item))
	}

	// The labels must have been created on the fly.
	response, err := server.ListLabels(authenticatedContext(), &proto.ListLabelsRequest{})
	if err != nil {
		t.Fatalf("failed to list the labels: %v", err)
	}
	if len(response.GetLabels()) != 2 {
		t.Errorf("expected 2 labels but got %d", len(response.GetLabels()))
	}
}

func TestListItemsFiltersByLabel(t *testing.T) {
	server := setupItemServer(t)
	// Items are created untriaged; triage them so they appear in the default
	// active listing filtered by label below.
	if _, err := server.CreateItem(authenticatedContext(), &proto.CreateItemRequest{
		Title:  "a",
		Labels: []string{"work", "urgent"},
	}); err != nil {
		t.Fatalf("failed to create the item: %v", err)
	}
	if _, err := server.CreateItem(authenticatedContext(), &proto.CreateItemRequest{
		Title:  "b",
		Labels: []string{"work"},
	}); err != nil {
		t.Fatalf("failed to create the item: %v", err)
	}
	// Triage both items to the bottom so they join the manual ordering.
	for _, title := range []string{"a", "b"} {
		listed, err := server.ListItems(authenticatedContext(), &proto.ListItemsRequest{View: proto.ItemView_ITEM_VIEW_UNTRIAGED})
		if err != nil {
			t.Fatalf("failed to list the untriaged items: %v", err)
		}
		var id uint32
		for _, item := range listed.GetActive() {
			if item.GetTitle() == title {
				id = item.GetId()
				break
			}
		}
		if id == 0 {
			t.Fatalf("failed to find the untriaged item %q", title)
		}
		if _, err := server.MoveItem(authenticatedContext(), &proto.MoveItemRequest{
			Id:     id,
			Anchor: &proto.MoveItemRequest_Bottom{Bottom: true},
		}); err != nil {
			t.Fatalf("failed to triage %q: %v", title, err)
		}
	}

	tests := []struct {
		name     string
		labels   []string
		expected []string
	}{
		{"no filter", nil, []string{"a", "b"}},
		{"one label", []string{"work"}, []string{"a", "b"}},
		{"two labels are combined with AND", []string{"work", "urgent"}, []string{"a"}},
		{"unknown label", []string{"nonexistent"}, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, err := server.ListItems(authenticatedContext(), &proto.ListItemsRequest{Labels: test.labels})
			if err != nil {
				t.Fatalf("expected no error but got %v", err)
			}

			titles := make([]string, 0, len(response.GetActive()))
			for _, item := range response.GetActive() {
				titles = append(titles, item.GetTitle())
			}
			if !equalStrings(titles, test.expected) {
				t.Errorf("expected %v but got %v", test.expected, titles)
			}
		})
	}
}

func TestListLabelsRejectsUnauthenticated(t *testing.T) {
	server := setupItemServer(t)
	_, err := server.ListLabels(context.Background(), &proto.ListLabelsRequest{})
	if got := status.Code(err); got != codes.Unauthenticated {
		t.Errorf("expected %v but got %v (%v)", codes.Unauthenticated, got, err)
	}
}

func TestCreateLabelRejectsUnauthenticated(t *testing.T) {
	server := setupItemServer(t)
	_, err := server.CreateLabel(context.Background(), &proto.CreateLabelRequest{Name: "work"})
	if got := status.Code(err); got != codes.Unauthenticated {
		t.Errorf("expected %v but got %v (%v)", codes.Unauthenticated, got, err)
	}
}

func TestUpdateItemLabelsRejectsUnauthenticated(t *testing.T) {
	server := setupItemServer(t)
	// The id and add/remove checks run before the auth check.
	_, err := server.UpdateItemLabels(context.Background(), &proto.UpdateItemLabelsRequest{
		Id:  1,
		Add: []string{"work"},
	})
	if got := status.Code(err); got != codes.Unauthenticated {
		t.Errorf("expected %v but got %v (%v)", codes.Unauthenticated, got, err)
	}
}

func TestUpdateItemLabelsAddsAndRemoves(t *testing.T) {
	server := setupItemServer(t)
	ids := createItems(t, server, "a")

	// Attach two labels, then in a second call add one and remove the other so
	// both branches run in the same transaction.
	if _, err := server.UpdateItemLabels(authenticatedContext(), &proto.UpdateItemLabelsRequest{
		Id:  ids["a"],
		Add: []string{"work", "urgent"},
	}); err != nil {
		t.Fatalf("failed to attach the labels: %v", err)
	}

	updated, err := server.UpdateItemLabels(authenticatedContext(), &proto.UpdateItemLabelsRequest{
		Id:     ids["a"],
		Add:    []string{"home"},
		Remove: []string{"urgent"},
	})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if !containsAll(itemLabelNames(updated), "work", "home") {
		t.Errorf("expected [work home] but got %v", itemLabelNames(updated))
	}
	if containsAll(itemLabelNames(updated), "urgent") {
		t.Errorf("expected urgent to have been removed but got %v", itemLabelNames(updated))
	}
}
