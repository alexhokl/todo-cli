package internal

import (
	"context"

	"github.com/alexhokl/todo-cli/database"
	"github.com/alexhokl/todo-cli/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ListLabels returns every known label owned by the caller, ordered by name.
func (s *TodoServer) ListLabels(ctx context.Context, _ *proto.ListLabelsRequest) (*proto.ListLabelsResponse, error) {
	ctx, span := startSpan(ctx, "ListLabels")
	defer span.End()

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	labels, err := database.ListLabels(s.DB.WithContext(ctx), userID)
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	result, err := toProtoLabels(labels)
	if err != nil {
		return nil, err
	}

	endSpanOk(span)

	return &proto.ListLabelsResponse{Labels: result}, nil
}

// CreateLabel creates a label explicitly.
func (s *TodoServer) CreateLabel(ctx context.Context, req *proto.CreateLabelRequest) (*proto.Label, error) {
	ctx, span := startSpan(ctx, "CreateLabel")
	defer span.End()

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	label, err := database.CreateLabel(s.DB.WithContext(ctx), userID, req.GetName())
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	result, err := toProtoLabel(*label)
	if err != nil {
		return nil, err
	}

	endSpanOk(span)

	return result, nil
}

// RenameLabel changes the name of an existing label.
func (s *TodoServer) RenameLabel(ctx context.Context, req *proto.RenameLabelRequest) (*proto.Label, error) {
	ctx, span := startSpan(ctx, "RenameLabel")
	defer span.End()

	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	label, err := database.RenameLabel(s.DB.WithContext(ctx), userID, uint(req.GetId()), req.GetName())
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	result, err := toProtoLabel(*label)
	if err != nil {
		return nil, err
	}

	endSpanOk(span)

	return result, nil
}

// DeleteLabel removes a label that is no longer attached to any todo.
func (s *TodoServer) DeleteLabel(ctx context.Context, req *proto.DeleteLabelRequest) (*emptypb.Empty, error) {
	ctx, span := startSpan(ctx, "DeleteLabel")
	defer span.End()

	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := database.DeleteLabel(s.DB.WithContext(ctx), userID, uint(req.GetId())); err != nil {
		return nil, mapDatabaseError(err)
	}

	endSpanOk(span)

	return &emptypb.Empty{}, nil
}

// UpdateTodoLabels attaches and detaches labels on a todo.
func (s *TodoServer) UpdateTodoLabels(ctx context.Context, req *proto.UpdateTodoLabelsRequest) (*proto.Todo, error) {
	ctx, span := startSpan(ctx, "UpdateTodoLabels")
	defer span.End()

	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if len(req.GetAdd()) == 0 && len(req.GetRemove()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "either add or remove is required")
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	todo, err := database.UpdateTodoLabels(
		s.DB.WithContext(ctx),
		userID,
		uint(req.GetId()),
		req.GetAdd(),
		req.GetRemove(),
	)
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	result, err := toProtoTodo(*todo)
	if err != nil {
		return nil, err
	}

	endSpanOk(span)

	return result, nil
}

func toProtoLabels(labels []database.Label) ([]*proto.Label, error) {
	result := make([]*proto.Label, 0, len(labels))
	for _, label := range labels {
		converted, err := toProtoLabel(label)
		if err != nil {
			return nil, err
		}
		result = append(result, converted)
	}

	return result, nil
}

func toProtoLabel(label database.Label) (*proto.Label, error) {
	id, err := toProtoID(label.ID)
	if err != nil {
		return nil, err
	}

	return &proto.Label{Id: id, Name: label.Name}, nil
}
