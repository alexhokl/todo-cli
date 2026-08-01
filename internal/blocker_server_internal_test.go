package internal

import (
	"context"
	"testing"

	"github.com/alexhokl/todo-cli/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateBlocker(t *testing.T) {
	server := setupItemServer(t)
	ids := createItems(t, server, "a")

	blocker, err := server.CreateBlocker(authenticatedContext(), &proto.CreateBlockerRequest{
		ItemId:      ids["a"],
		Description: "  waiting on legal  ",
	})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if blocker.GetDescription() != "waiting on legal" {
		t.Errorf("expected the description to be trimmed but got %q", blocker.GetDescription())
	}

	// An empty description is rejected.
	_, err = server.CreateBlocker(authenticatedContext(), &proto.CreateBlockerRequest{
		ItemId:      ids["a"],
		Description: "   ",
	})
	if got := status.Code(err); got != codes.InvalidArgument {
		t.Errorf("expected %v but got %v (%v)", codes.InvalidArgument, got, err)
	}
}

func TestListBlockers(t *testing.T) {
	server := setupItemServer(t)
	ids := createItems(t, server, "a")

	for _, desc := range []string{"first", "second"} {
		if _, err := server.CreateBlocker(authenticatedContext(), &proto.CreateBlockerRequest{
			ItemId:      ids["a"],
			Description: desc,
		}); err != nil {
			t.Fatalf("failed to create the blocker: %v", err)
		}
	}

	response, err := server.ListBlockers(authenticatedContext(), &proto.ListBlockersRequest{ItemId: ids["a"]})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if len(response.GetBlockers()) != 2 {
		t.Fatalf("expected 2 blockers but got %d", len(response.GetBlockers()))
	}
	if response.GetBlockers()[0].GetDescription() != "first" {
		t.Errorf("expected the blockers to be ordered by id but got %v", response.GetBlockers())
	}
}

func TestUpdateBlocker(t *testing.T) {
	server := setupItemServer(t)
	ids := createItems(t, server, "a")
	blocker, err := server.CreateBlocker(authenticatedContext(), &proto.CreateBlockerRequest{
		ItemId:      ids["a"],
		Description: "old reason",
	})
	if err != nil {
		t.Fatalf("failed to create the blocker: %v", err)
	}

	updated, err := server.UpdateBlocker(authenticatedContext(), &proto.UpdateBlockerRequest{
		Id:          blocker.GetId(),
		Description: "  new reason  ",
	})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if updated.GetDescription() != "new reason" {
		t.Errorf("expected the description to be trimmed but got %q", updated.GetDescription())
	}

	// An empty description is rejected.
	emptyReq := &proto.UpdateBlockerRequest{
		Id:          blocker.GetId(),
		Description: "  ",
	}
	_, err = server.UpdateBlocker(authenticatedContext(), emptyReq)
	if got := status.Code(err); got != codes.InvalidArgument {
		t.Errorf("expected %v but got %v (%v)", codes.InvalidArgument, got, err)
	}
}

func TestDeleteBlocker(t *testing.T) {
	server := setupItemServer(t)
	ids := createItems(t, server, "a")
	blocker, err := server.CreateBlocker(authenticatedContext(), &proto.CreateBlockerRequest{
		ItemId:      ids["a"],
		Description: "blocked",
	})
	if err != nil {
		t.Fatalf("failed to create the blocker: %v", err)
	}

	if _, err := server.DeleteBlocker(authenticatedContext(), &proto.DeleteBlockerRequest{Id: blocker.GetId()}); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	// A second delete is reported as not found.
	_, err = server.DeleteBlocker(authenticatedContext(), &proto.DeleteBlockerRequest{Id: blocker.GetId()})
	if got := status.Code(err); got != codes.NotFound {
		t.Errorf("expected %v but got %v (%v)", codes.NotFound, got, err)
	}
}

func TestBlockerErrorCodes(t *testing.T) {
	server := setupItemServer(t)
	ids := createItems(t, server, "a")

	tests := []struct {
		name     string
		request  func(ids map[string]uint32) *proto.DeleteBlockerRequest
		expected codes.Code
	}{
		{
			name: "missing id",
			request: func(_ map[string]uint32) *proto.DeleteBlockerRequest {
				return &proto.DeleteBlockerRequest{}
			},
			expected: codes.InvalidArgument,
		},
		{
			name: "unknown id",
			request: func(_ map[string]uint32) *proto.DeleteBlockerRequest {
				return &proto.DeleteBlockerRequest{Id: 404}
			},
			expected: codes.NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := server.DeleteBlocker(authenticatedContext(), test.request(ids))
			if got := status.Code(err); got != test.expected {
				t.Errorf("expected %v but got %v (%v)", test.expected, got, err)
			}
		})
	}
}

func TestListBlockersRejectsUnauthenticated(t *testing.T) {
	server := setupItemServer(t)
	_, err := server.ListBlockers(context.Background(), &proto.ListBlockersRequest{ItemId: 1})
	if got := status.Code(err); got != codes.Unauthenticated {
		t.Errorf("expected %v but got %v (%v)", codes.Unauthenticated, got, err)
	}
}

func TestCreateBlockerRejectsUnauthenticated(t *testing.T) {
	server := setupItemServer(t)
	_, err := server.CreateBlocker(context.Background(), &proto.CreateBlockerRequest{ItemId: 1, Description: "x"})
	if got := status.Code(err); got != codes.Unauthenticated {
		t.Errorf("expected %v but got %v (%v)", codes.Unauthenticated, got, err)
	}
}