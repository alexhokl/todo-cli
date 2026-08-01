package internal

import (
	"context"

	"github.com/alexhokl/todo-cli/database"
	"github.com/alexhokl/todo-cli/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ListEfforts returns every known effort owned by the caller, ordered by name.
func (s *ItemServer) ListEfforts(ctx context.Context, _ *proto.ListEffortsRequest) (*proto.ListEffortsResponse, error) {
	ctx, span := startSpan(ctx, "ListEfforts")
	defer span.End()

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	efforts, err := database.ListEfforts(s.DB.WithContext(ctx), userID)
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	result, err := toProtoEfforts(efforts)
	if err != nil {
		return nil, err
	}

	endSpanOk(span)

	return &proto.ListEffortsResponse{Efforts: result}, nil
}

// CreateEffort creates an effort explicitly.
func (s *ItemServer) CreateEffort(ctx context.Context, req *proto.CreateEffortRequest) (*proto.Effort, error) {
	ctx, span := startSpan(ctx, "CreateEffort")
	defer span.End()

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	effort, err := database.CreateEffort(s.DB.WithContext(ctx), userID, req.GetName())
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	result, err := toProtoEffort(*effort)
	if err != nil {
		return nil, err
	}

	endSpanOk(span)

	return result, nil
}

// RenameEffort changes the name of an existing effort.
func (s *ItemServer) RenameEffort(ctx context.Context, req *proto.RenameEffortRequest) (*proto.Effort, error) {
	ctx, span := startSpan(ctx, "RenameEffort")
	defer span.End()

	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	effort, err := database.RenameEffort(s.DB.WithContext(ctx), userID, uint(req.GetId()), req.GetName())
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	result, err := toProtoEffort(*effort)
	if err != nil {
		return nil, err
	}

	endSpanOk(span)

	return result, nil
}

// DeleteEffort removes an effort that is no longer attached to any item.
func (s *ItemServer) DeleteEffort(ctx context.Context, req *proto.DeleteEffortRequest) (*emptypb.Empty, error) {
	ctx, span := startSpan(ctx, "DeleteEffort")
	defer span.End()

	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := database.DeleteEffort(s.DB.WithContext(ctx), userID, uint(req.GetId())); err != nil {
		return nil, mapDatabaseError(err)
	}

	endSpanOk(span)

	return &emptypb.Empty{}, nil
}

// SetItemEffort attaches an effort to an item by name, or clears it when the
// name is empty.
func (s *ItemServer) SetItemEffort(ctx context.Context, req *proto.SetItemEffortRequest) (*proto.Item, error) {
	ctx, span := startSpan(ctx, "SetItemEffort")
	defer span.End()

	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	item, err := database.SetItemEffort(s.DB.WithContext(ctx), userID, uint(req.GetId()), req.GetEffort())
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	result, err := toProtoItem(*item)
	if err != nil {
		return nil, err
	}

	endSpanOk(span)

	return result, nil
}

func toProtoEfforts(efforts []database.Effort) ([]*proto.Effort, error) {
	result := make([]*proto.Effort, 0, len(efforts))
	for _, effort := range efforts {
		converted, err := toProtoEffort(effort)
		if err != nil {
			return nil, err
		}
		result = append(result, converted)
	}

	return result, nil
}

func toProtoEffort(effort database.Effort) (*proto.Effort, error) {
	id, err := toProtoID(effort.ID)
	if err != nil {
		return nil, err
	}

	return &proto.Effort{Id: id, Name: effort.Name}, nil
}