package internal

import (
	"context"
	"testing"

	"github.com/alexhokl/todo-cli/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateEffortNormalisesAndRejectsDuplicates(t *testing.T) {
	server := setupItemServer(t)

	effort, err := server.CreateEffort(authenticatedContext(), &proto.CreateEffortRequest{Name: "  High  "})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if effort.GetName() != "high" {
		t.Errorf("expected %q but got %q", "high", effort.GetName())
	}

	_, err = server.CreateEffort(authenticatedContext(), &proto.CreateEffortRequest{Name: "HIGH"})
	if got := status.Code(err); got != codes.AlreadyExists {
		t.Errorf("expected %v but got %v (%v)", codes.AlreadyExists, got, err)
	}

	_, err = server.CreateEffort(authenticatedContext(), &proto.CreateEffortRequest{Name: "   "})
	if got := status.Code(err); got != codes.InvalidArgument {
		t.Errorf("expected %v but got %v (%v)", codes.InvalidArgument, got, err)
	}
}

func TestListEfforts(t *testing.T) {
	server := setupItemServer(t)
	for _, name := range []string{"high", "low"} {
		if _, err := server.CreateEffort(authenticatedContext(), &proto.CreateEffortRequest{Name: name}); err != nil {
			t.Fatalf("failed to create the effort %q: %v", name, err)
		}
	}

	response, err := server.ListEfforts(authenticatedContext(), &proto.ListEffortsRequest{})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if len(response.GetEfforts()) != 2 {
		t.Fatalf("expected 2 efforts but got %d", len(response.GetEfforts()))
	}
	if response.GetEfforts()[0].GetName() != "high" {
		t.Errorf("expected the efforts to be ordered by name but got %v", response.GetEfforts())
	}
}

func TestRenameEffortErrorCodes(t *testing.T) {
	server := setupItemServer(t)
	high, err := server.CreateEffort(authenticatedContext(), &proto.CreateEffortRequest{Name: "high"})
	if err != nil {
		t.Fatalf("failed to create the effort: %v", err)
	}
	if _, err := server.CreateEffort(authenticatedContext(), &proto.CreateEffortRequest{Name: "low"}); err != nil {
		t.Fatalf("failed to create the effort: %v", err)
	}

	tests := []struct {
		name     string
		id       uint32
		newName  string
		expected codes.Code
	}{
		{"missing id", 0, "medium", codes.InvalidArgument},
		{"unknown id", 404, "medium", codes.NotFound},
		{"empty name", high.GetId(), "  ", codes.InvalidArgument},
		{"name already taken", high.GetId(), "low", codes.AlreadyExists},
		{"valid", high.GetId(), "medium", codes.OK},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := &proto.RenameEffortRequest{Id: test.id, Name: test.newName}
			_, err := server.RenameEffort(authenticatedContext(), req)
			if got := status.Code(err); got != test.expected {
				t.Errorf("expected %v but got %v (%v)", test.expected, got, err)
			}
		})
	}
}

func TestDeleteEffortRefusesWhileInUse(t *testing.T) {
	server := setupItemServer(t)
	ids := createItems(t, server, "a")
	effort, err := server.CreateEffort(authenticatedContext(), &proto.CreateEffortRequest{Name: "high"})
	if err != nil {
		t.Fatalf("failed to create the effort: %v", err)
	}

	tagged, err := server.SetItemEffort(authenticatedContext(), &proto.SetItemEffortRequest{
		Id:     ids["a"],
		Effort: "high",
	})
	if err != nil {
		t.Fatalf("failed to set the effort: %v", err)
	}
	if tagged.GetEffort() == nil || tagged.GetEffort().GetName() != "high" {
		t.Errorf("expected effort high but got %v", tagged.GetEffort())
	}

	_, err = server.DeleteEffort(authenticatedContext(), &proto.DeleteEffortRequest{Id: effort.GetId()})
	if got := status.Code(err); got != codes.FailedPrecondition {
		t.Errorf("expected %v but got %v (%v)", codes.FailedPrecondition, got, err)
	}

	// Clearing the effort allows the delete.
	if _, err := server.SetItemEffort(authenticatedContext(), &proto.SetItemEffortRequest{
		Id:     ids["a"],
		Effort: "",
	}); err != nil {
		t.Fatalf("failed to clear the effort: %v", err)
	}

	if _, err := server.DeleteEffort(authenticatedContext(), &proto.DeleteEffortRequest{Id: effort.GetId()}); err != nil {
		t.Errorf("expected no error but got %v", err)
	}
}

func TestDeleteEffortErrorCodes(t *testing.T) {
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
			_, err := server.DeleteEffort(authenticatedContext(), &proto.DeleteEffortRequest{Id: test.id})
			if got := status.Code(err); got != test.expected {
				t.Errorf("expected %v but got %v (%v)", test.expected, got, err)
			}
		})
	}
}

func TestSetItemEffortErrorCodes(t *testing.T) {
	server := setupItemServer(t)
	ids := createItems(t, server, "a")

	tests := []struct {
		name     string
		request  func(ids map[string]uint32) *proto.SetItemEffortRequest
		expected codes.Code
	}{
		{
			name: "missing id",
			request: func(_ map[string]uint32) *proto.SetItemEffortRequest {
				return &proto.SetItemEffortRequest{Effort: "high"}
			},
			expected: codes.InvalidArgument,
		},
		{
			name: "unknown item",
			request: func(_ map[string]uint32) *proto.SetItemEffortRequest {
				return &proto.SetItemEffortRequest{Id: 404, Effort: "high"}
			},
			expected: codes.NotFound,
		},
		{
			name: "unknown effort",
			request: func(ids map[string]uint32) *proto.SetItemEffortRequest {
				return &proto.SetItemEffortRequest{Id: ids["a"], Effort: "unknown"}
			},
			expected: codes.NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := server.SetItemEffort(authenticatedContext(), test.request(ids))
			if got := status.Code(err); got != test.expected {
				t.Errorf("expected %v but got %v (%v)", test.expected, got, err)
			}
		})
	}
}

func TestSetItemEffortSetsAndClears(t *testing.T) {
	server := setupItemServer(t)
	ids := createItems(t, server, "a")

	if _, err := server.CreateEffort(authenticatedContext(), &proto.CreateEffortRequest{Name: "high"}); err != nil {
		t.Fatalf("failed to create the effort: %v", err)
	}

	set, err := server.SetItemEffort(authenticatedContext(), &proto.SetItemEffortRequest{
		Id:     ids["a"],
		Effort: "High",
	})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if set.GetEffort() == nil || set.GetEffort().GetName() != "high" {
		t.Errorf("expected effort high but got %v", set.GetEffort())
	}

	cleared, err := server.SetItemEffort(authenticatedContext(), &proto.SetItemEffortRequest{
		Id:     ids["a"],
		Effort: "  ",
	})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if cleared.GetEffort() != nil {
		t.Errorf("expected the effort to be cleared but got %v", cleared.GetEffort())
	}
}

func TestListEffortsRejectsUnauthenticated(t *testing.T) {
	server := setupItemServer(t)
	_, err := server.ListEfforts(context.Background(), &proto.ListEffortsRequest{})
	if got := status.Code(err); got != codes.Unauthenticated {
		t.Errorf("expected %v but got %v (%v)", codes.Unauthenticated, got, err)
	}
}

func TestCreateEffortRejectsUnauthenticated(t *testing.T) {
	server := setupItemServer(t)
	_, err := server.CreateEffort(context.Background(), &proto.CreateEffortRequest{Name: "high"})
	if got := status.Code(err); got != codes.Unauthenticated {
		t.Errorf("expected %v but got %v (%v)", codes.Unauthenticated, got, err)
	}
}

func TestSetItemEffortRejectsUnauthenticated(t *testing.T) {
	server := setupItemServer(t)
	_, err := server.SetItemEffort(context.Background(), &proto.SetItemEffortRequest{Id: 1, Effort: "high"})
	if got := status.Code(err); got != codes.Unauthenticated {
		t.Errorf("expected %v but got %v (%v)", codes.Unauthenticated, got, err)
	}
}