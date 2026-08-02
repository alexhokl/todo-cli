package internal

import (
	"context"

	"github.com/alexhokl/todo-cli/database"
	"github.com/alexhokl/todo-cli/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ListComments returns every comment attached to the given item.
func (s *ItemServer) ListComments(ctx context.Context, req *proto.ListCommentsRequest) (*proto.ListCommentsResponse, error) {
	ctx, span := startSpan(ctx, "ListComments")
	defer span.End()

	if req.GetItemId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "item_id is required")
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	comments, err := database.ListComments(s.DB.WithContext(ctx), userID, uint(req.GetItemId()))
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	result, err := toProtoComments(comments)
	if err != nil {
		return nil, err
	}

	endSpanOk(span)

	return &proto.ListCommentsResponse{Comments: result}, nil
}

// CreateComment attaches a new comment to an item.
func (s *ItemServer) CreateComment(ctx context.Context, req *proto.CreateCommentRequest) (*proto.Comment, error) {
	ctx, span := startSpan(ctx, "CreateComment")
	defer span.End()

	if req.GetItemId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "item_id is required")
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	comment, err := database.CreateComment(s.DB.WithContext(ctx), userID, uint(req.GetItemId()), req.GetBody())
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	result, err := toProtoComment(*comment)
	if err != nil {
		return nil, err
	}

	endSpanOk(span)

	return result, nil
}

// UpdateComment edits the body of an existing comment.
func (s *ItemServer) UpdateComment(ctx context.Context, req *proto.UpdateCommentRequest) (*proto.Comment, error) {
	ctx, span := startSpan(ctx, "UpdateComment")
	defer span.End()

	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	comment, err := database.UpdateComment(s.DB.WithContext(ctx), userID, uint(req.GetId()), req.GetBody())
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	result, err := toProtoComment(*comment)
	if err != nil {
		return nil, err
	}

	endSpanOk(span)

	return result, nil
}

// DeleteComment removes a comment.
func (s *ItemServer) DeleteComment(ctx context.Context, req *proto.DeleteCommentRequest) (*emptypb.Empty, error) {
	ctx, span := startSpan(ctx, "DeleteComment")
	defer span.End()

	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := database.DeleteComment(s.DB.WithContext(ctx), userID, uint(req.GetId())); err != nil {
		return nil, mapDatabaseError(err)
	}

	endSpanOk(span)

	return &emptypb.Empty{}, nil
}

func toProtoComments(comments []database.Comment) ([]*proto.Comment, error) {
	result := make([]*proto.Comment, 0, len(comments))
	for _, comment := range comments {
		converted, err := toProtoComment(comment)
		if err != nil {
			return nil, err
		}
		result = append(result, converted)
	}

	return result, nil
}

func toProtoComment(comment database.Comment) (*proto.Comment, error) {
	id, err := toProtoID(comment.ID)
	if err != nil {
		return nil, err
	}

	return &proto.Comment{
		Id:        id,
		Body:      comment.Body,
		CreatedAt: timestamppb.New(comment.CreatedAt),
		Author:    comment.User.Username,
	}, nil
}