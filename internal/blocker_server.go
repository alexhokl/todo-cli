package internal

import (
	"context"

	"github.com/alexhokl/todo-cli/database"
	"github.com/alexhokl/todo-cli/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ListBlockers returns every blocker attached to the given item.
func (s *ItemServer) ListBlockers(ctx context.Context, req *proto.ListBlockersRequest) (*proto.ListBlockersResponse, error) {
	ctx, span := startSpan(ctx, "ListBlockers")
	defer span.End()

	if req.GetItemId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "item_id is required")
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	blockers, err := database.ListBlockers(s.DB.WithContext(ctx), userID, uint(req.GetItemId()))
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	result, err := toProtoBlockers(blockers)
	if err != nil {
		return nil, err
	}

	endSpanOk(span)

	return &proto.ListBlockersResponse{Blockers: result}, nil
}

// CreateBlocker attaches a new blocker to an item.
func (s *ItemServer) CreateBlocker(ctx context.Context, req *proto.CreateBlockerRequest) (*proto.Blocker, error) {
	ctx, span := startSpan(ctx, "CreateBlocker")
	defer span.End()

	if req.GetItemId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "item_id is required")
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	blocker, err := database.CreateBlocker(s.DB.WithContext(ctx), userID, uint(req.GetItemId()), req.GetDescription())
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	result, err := toProtoBlocker(*blocker)
	if err != nil {
		return nil, err
	}

	endSpanOk(span)

	return result, nil
}

// UpdateBlocker changes the description of an existing blocker.
func (s *ItemServer) UpdateBlocker(ctx context.Context, req *proto.UpdateBlockerRequest) (*proto.Blocker, error) {
	ctx, span := startSpan(ctx, "UpdateBlocker")
	defer span.End()

	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	blocker, err := database.UpdateBlocker(s.DB.WithContext(ctx), userID, uint(req.GetId()), req.GetDescription())
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	result, err := toProtoBlocker(*blocker)
	if err != nil {
		return nil, err
	}

	endSpanOk(span)

	return result, nil
}

// DeleteBlocker removes a blocker.
func (s *ItemServer) DeleteBlocker(ctx context.Context, req *proto.DeleteBlockerRequest) (*emptypb.Empty, error) {
	ctx, span := startSpan(ctx, "DeleteBlocker")
	defer span.End()

	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := database.DeleteBlocker(s.DB.WithContext(ctx), userID, uint(req.GetId())); err != nil {
		return nil, mapDatabaseError(err)
	}

	endSpanOk(span)

	return &emptypb.Empty{}, nil
}

func toProtoBlockers(blockers []database.Blocker) ([]*proto.Blocker, error) {
	result := make([]*proto.Blocker, 0, len(blockers))
	for _, blocker := range blockers {
		converted, err := toProtoBlocker(blocker)
		if err != nil {
			return nil, err
		}
		result = append(result, converted)
	}

	return result, nil
}

func toProtoBlocker(blocker database.Blocker) (*proto.Blocker, error) {
	id, err := toProtoID(blocker.ID)
	if err != nil {
		return nil, err
	}

	return &proto.Blocker{Id: id, Description: blocker.Description}, nil
}