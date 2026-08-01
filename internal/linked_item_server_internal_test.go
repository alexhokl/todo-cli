package internal

import (
	"context"
	"testing"

	"github.com/alexhokl/todo-cli/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUpdateItemLinksSymmetric(t *testing.T) {
	server := setupItemServer(t)
	ids := createItems(t, server, "a", "b", "c")

	updated, err := server.UpdateItemLinks(authenticatedContext(), &proto.UpdateItemLinksRequest{
		Id:  ids["a"],
		Add: []uint32{ids["b"]},
	})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if len(updated.GetLinkedItems()) != 1 || updated.GetLinkedItems()[0].GetId() != ids["b"] {
		t.Errorf("expected a to link to b but got %v", updated.GetLinkedItems())
	}

	// The link is symmetric: b also links to a (verified via the list, which
	// preloads LinkedItems on every item).
	listed, err := server.ListItems(authenticatedContext(), &proto.ListItemsRequest{})
	if err != nil {
		t.Fatalf("failed to list items: %v", err)
	}
	for _, item := range listed.GetActive() {
		if item.GetId() == ids["b"] {
			if len(item.GetLinkedItems()) != 1 || item.GetLinkedItems()[0].GetId() != ids["a"] {
				t.Errorf("expected b to link back to a (symmetric) but got %v", item.GetLinkedItems())
			}
		}
	}
}

func TestUpdateItemLinksErrorCodes(t *testing.T) {
	server := setupItemServer(t)
	ids := createItems(t, server, "a", "b")

	tests := []struct {
		name     string
		request  func(ids map[string]uint32) *proto.UpdateItemLinksRequest
		expected codes.Code
	}{
		{
			name: "missing id",
			request: func(_ map[string]uint32) *proto.UpdateItemLinksRequest {
				return &proto.UpdateItemLinksRequest{Add: []uint32{1}}
			},
			expected: codes.InvalidArgument,
		},
		{
			name: "neither add nor remove",
			request: func(ids map[string]uint32) *proto.UpdateItemLinksRequest {
				return &proto.UpdateItemLinksRequest{Id: ids["a"]}
			},
			expected: codes.InvalidArgument,
		},
		{
			name: "self-link",
			request: func(ids map[string]uint32) *proto.UpdateItemLinksRequest {
				return &proto.UpdateItemLinksRequest{Id: ids["a"], Add: []uint32{ids["a"]}}
			},
			expected: codes.InvalidArgument,
		},
		{
			name: "unknown target",
			request: func(ids map[string]uint32) *proto.UpdateItemLinksRequest {
				return &proto.UpdateItemLinksRequest{Id: ids["a"], Add: []uint32{404}}
			},
			expected: codes.NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := server.UpdateItemLinks(authenticatedContext(), test.request(ids))
			if got := status.Code(err); got != test.expected {
				t.Errorf("expected %v but got %v (%v)", test.expected, got, err)
			}
		})
	}
}

func TestCreateItemWithLinks(t *testing.T) {
	server := setupItemServer(t)
	// Create target items first.
	targets := createItems(t, server, "t1", "t2")

	// Create a new item linked to both targets.
	item, err := server.CreateItem(authenticatedContext(), &proto.CreateItemRequest{
		Title:       "linked",
		LinkItemIds: []uint32{targets["t1"], targets["t2"]},
	})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if len(item.GetLinkedItems()) != 2 {
		t.Errorf("expected 2 linked items but got %d", len(item.GetLinkedItems()))
	}

	// The links are symmetric: the targets link back.
	for _, title := range []string{"t1", "t2"} {
		listed, err := server.ListItems(authenticatedContext(), &proto.ListItemsRequest{})
		if err != nil {
			t.Fatalf("failed to list items: %v", err)
		}
		for _, it := range listed.GetActive() {
			if it.GetTitle() == title {
				if len(it.GetLinkedItems()) != 1 || it.GetLinkedItems()[0].GetTitle() != "linked" {
					t.Errorf("expected %q to link back to linked but got %v", title, it.GetLinkedItems())
				}
			}
		}
	}
}

func TestUpdateItemLinksRejectsUnauthenticated(t *testing.T) {
	server := setupItemServer(t)
	_, err := server.UpdateItemLinks(context.Background(), &proto.UpdateItemLinksRequest{Id: 1, Add: []uint32{2}})
	if got := status.Code(err); got != codes.Unauthenticated {
		t.Errorf("expected %v but got %v (%v)", codes.Unauthenticated, got, err)
	}
}