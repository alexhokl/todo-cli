package internal

import (
	"context"
	"testing"

	"github.com/alexhokl/todo-cli/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateComment(t *testing.T) {
	server := setupItemServer(t)
	ids := createItems(t, server, "a")

	comment, err := server.CreateComment(authenticatedContext(), &proto.CreateCommentRequest{
		ItemId: ids["a"],
		Body:   "  needs review  ",
	})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if comment.GetBody() != "needs review" {
		t.Errorf("expected the body to be trimmed but got %q", comment.GetBody())
	}
	if comment.GetAuthor() != "testuser" {
		t.Errorf("expected the author to be testuser but got %q", comment.GetAuthor())
	}
	if comment.GetCreatedAt() == nil {
		t.Errorf("expected the created_at timestamp to be set")
	}

	// An empty body is rejected.
	_, err = server.CreateComment(authenticatedContext(), &proto.CreateCommentRequest{
		ItemId: ids["a"],
		Body:   "   ",
	})
	if got := status.Code(err); got != codes.InvalidArgument {
		t.Errorf("expected %v but got %v (%v)", codes.InvalidArgument, got, err)
	}
}

func TestListComments(t *testing.T) {
	server := setupItemServer(t)
	ids := createItems(t, server, "a")

	for _, body := range []string{"first", "second"} {
		if _, err := server.CreateComment(authenticatedContext(), &proto.CreateCommentRequest{
			ItemId: ids["a"],
			Body:   body,
		}); err != nil {
			t.Fatalf("failed to create the comment: %v", err)
		}
	}

	response, err := server.ListComments(authenticatedContext(), &proto.ListCommentsRequest{ItemId: ids["a"]})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if len(response.GetComments()) != 2 {
		t.Fatalf("expected 2 comments but got %d", len(response.GetComments()))
	}
	if response.GetComments()[0].GetBody() != "first" {
		t.Errorf("expected the comments to be ordered by id but got %v", response.GetComments())
	}
	if response.GetComments()[0].GetAuthor() != "testuser" {
		t.Errorf("expected the author to be testuser but got %q", response.GetComments()[0].GetAuthor())
	}
}

func TestUpdateComment(t *testing.T) {
	server := setupItemServer(t)
	ids := createItems(t, server, "a")
	comment, err := server.CreateComment(authenticatedContext(), &proto.CreateCommentRequest{
		ItemId: ids["a"],
		Body:   "old remark",
	})
	if err != nil {
		t.Fatalf("failed to create the comment: %v", err)
	}

	updated, err := server.UpdateComment(authenticatedContext(), &proto.UpdateCommentRequest{
		Id:   comment.GetId(),
		Body: "  new remark  ",
	})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if updated.GetBody() != "new remark" {
		t.Errorf("expected the body to be trimmed but got %q", updated.GetBody())
	}

	// An empty body is rejected.
	emptyReq := &proto.UpdateCommentRequest{
		Id:   comment.GetId(),
		Body: "  ",
	}
	_, err = server.UpdateComment(authenticatedContext(), emptyReq)
	if got := status.Code(err); got != codes.InvalidArgument {
		t.Errorf("expected %v but got %v (%v)", codes.InvalidArgument, got, err)
	}
}

func TestDeleteComment(t *testing.T) {
	server := setupItemServer(t)
	ids := createItems(t, server, "a")
	comment, err := server.CreateComment(authenticatedContext(), &proto.CreateCommentRequest{
		ItemId: ids["a"],
		Body:   "remark",
	})
	if err != nil {
		t.Fatalf("failed to create the comment: %v", err)
	}

	if _, err := server.DeleteComment(authenticatedContext(), &proto.DeleteCommentRequest{Id: comment.GetId()}); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	// A second delete is reported as not found.
	_, err = server.DeleteComment(authenticatedContext(), &proto.DeleteCommentRequest{Id: comment.GetId()})
	if got := status.Code(err); got != codes.NotFound {
		t.Errorf("expected %v but got %v (%v)", codes.NotFound, got, err)
	}
}

func TestCommentErrorCodes(t *testing.T) {
	server := setupItemServer(t)
	ids := createItems(t, server, "a")

	tests := []struct {
		name     string
		request  func(ids map[string]uint32) *proto.DeleteCommentRequest
		expected codes.Code
	}{
		{
			name: "missing id",
			request: func(_ map[string]uint32) *proto.DeleteCommentRequest {
				return &proto.DeleteCommentRequest{}
			},
			expected: codes.InvalidArgument,
		},
		{
			name: "unknown id",
			request: func(_ map[string]uint32) *proto.DeleteCommentRequest {
				return &proto.DeleteCommentRequest{Id: 404}
			},
			expected: codes.NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := server.DeleteComment(authenticatedContext(), test.request(ids))
			if got := status.Code(err); got != test.expected {
				t.Errorf("expected %v but got %v (%v)", test.expected, got, err)
			}
		})
	}
}

func TestListCommentsRejectsUnauthenticated(t *testing.T) {
	server := setupItemServer(t)
	_, err := server.ListComments(context.Background(), &proto.ListCommentsRequest{ItemId: 1})
	if got := status.Code(err); got != codes.Unauthenticated {
		t.Errorf("expected %v but got %v (%v)", codes.Unauthenticated, got, err)
	}
}

func TestCreateCommentRejectsUnauthenticated(t *testing.T) {
	server := setupItemServer(t)
	_, err := server.CreateComment(context.Background(), &proto.CreateCommentRequest{ItemId: 1, Body: "x"})
	if got := status.Code(err); got != codes.Unauthenticated {
		t.Errorf("expected %v but got %v (%v)", codes.Unauthenticated, got, err)
	}
}