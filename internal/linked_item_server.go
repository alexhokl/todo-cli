package internal

import (
	"context"

	"github.com/alexhokl/todo-cli/database"
	"github.com/alexhokl/todo-cli/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UpdateItemLinks attaches and detaches links between an item and other items.
// The relationship is symmetric: linking A to B also links B to A.
func (s *ItemServer) UpdateItemLinks(ctx context.Context, req *proto.UpdateItemLinksRequest) (*proto.Item, error) {
	ctx, span := startSpan(ctx, "UpdateItemLinks")
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

	addIDs := make([]uint, 0, len(req.GetAdd()))
	for _, id := range req.GetAdd() {
		addIDs = append(addIDs, uint(id))
	}
	removeIDs := make([]uint, 0, len(req.GetRemove()))
	for _, id := range req.GetRemove() {
		removeIDs = append(removeIDs, uint(id))
	}

	item, err := database.UpdateItemLinks(s.DB.WithContext(ctx), userID, uint(req.GetId()), addIDs, removeIDs)
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