package internal

import (
	"context"
	"errors"
	"math"

	"github.com/alexhokl/todo-cli/database"
	"github.com/alexhokl/todo-cli/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// ItemServer implements the item gRPC service on top of a GORM database.
type ItemServer struct {
	proto.UnimplementedItemServiceServer
	DB *gorm.DB
}

// NewItemServer creates an item service implementation backed by the given
// database connection.
func NewItemServer(db *gorm.DB) *ItemServer {
	return &ItemServer{DB: db}
}

// userIDFromContext extracts the authenticated user identifier placed in the
// context by the authentication interceptor. Handlers return an
// Unauthenticated status when it is missing so the request is rejected before
// any database access.
func userIDFromContext(ctx context.Context) (uint, error) {
	userID, ok := ctx.Value(contextKeyUser{}).(uint)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "authentication required")
	}
	return userID, nil
}

// ListItems returns the active items in manual order and the completed items
// ordered by how recently they were updated.
func (s *ItemServer) ListItems(ctx context.Context, req *proto.ListItemsRequest) (*proto.ListItemsResponse, error) {
	ctx, span := startSpan(ctx, "ListItems")
	defer span.End()

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	filter := database.ItemFilter{Labels: req.GetLabels()}

	active, err := database.ListActive(s.DB.WithContext(ctx), userID, filter)
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	completed, err := database.ListCompleted(s.DB.WithContext(ctx), userID, filter)
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	activeItems, err := toProtoItems(active)
	if err != nil {
		return nil, err
	}
	completedItems, err := toProtoItems(completed)
	if err != nil {
		return nil, err
	}

	endSpanOk(span)

	return &proto.ListItemsResponse{
		Active:    activeItems,
		Completed: completedItems,
	}, nil
}

// CreateItem appends a new item to the end of the manual order.
func (s *ItemServer) CreateItem(ctx context.Context, req *proto.CreateItemRequest) (*proto.Item, error) {
	ctx, span := startSpan(ctx, "CreateItem")
	defer span.End()

	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	item := database.Item{
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
		ListID:      optionalID(req.ListId),
	}
	if req.DueDate != nil {
		dueDate := req.GetDueDate().AsTime()
		item.DueDate = &dueDate
	}

	db := s.DB.WithContext(ctx)
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := database.AssignInitialPosition(tx, &item, userID); err != nil {
			return err
		}
		// Labels are resolved before the insert so that the item is created
		// with its join rows already in place.
		labels, err := database.FindOrCreateLabels(tx, userID, req.GetLabels())
		if err != nil {
			return err
		}
		item.Labels = labels

		return tx.Create(&item).Error
	})
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	result, err := toProtoItem(item)
	if err != nil {
		return nil, err
	}

	endSpanOk(span)

	return result, nil
}

// MoveItem places an item immediately before or after another item, optionally
// reassigning its list in the same operation.
func (s *ItemServer) MoveItem(ctx context.Context, req *proto.MoveItemRequest) (*proto.Item, error) {
	ctx, span := startSpan(ctx, "MoveItem")
	defer span.End()

	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var anchor database.MoveAnchor
	switch {
	case req.GetBeforeId() != 0:
		anchor = database.MoveAnchor{TargetID: uint(req.GetBeforeId()), Before: true}
	case req.GetAfterId() != 0:
		anchor = database.MoveAnchor{TargetID: uint(req.GetAfterId()), Before: false}
	default:
		return nil, status.Error(codes.InvalidArgument, "either before_id or after_id is required")
	}

	opts := database.MoveOptions{ChangeList: req.GetChangeList()}
	if opts.ChangeList {
		opts.ListID = optionalID(req.ListId)
	}

	item, err := database.MoveItem(s.DB.WithContext(ctx), userID, uint(req.GetId()), anchor, opts)
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

// SetItemDone completes or reopens an item.
func (s *ItemServer) SetItemDone(ctx context.Context, req *proto.SetItemDoneRequest) (*proto.Item, error) {
	ctx, span := startSpan(ctx, "SetItemDone")
	defer span.End()

	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	item, err := database.SetDone(s.DB.WithContext(ctx), userID, uint(req.GetId()), req.GetDone())
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

// mapDatabaseError translates the sentinel errors of the database package into
// gRPC status errors. The error logging interceptor takes care of recording
// them, so handlers only have to return the result.
func mapDatabaseError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, database.ErrItemNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, database.ErrLabelNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, database.ErrLabelExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, database.ErrLabelNameEmpty):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, database.ErrItemCompleted),
		errors.Is(err, database.ErrAnchorCompleted),
		errors.Is(err, database.ErrLabelInUse):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Errorf(codes.Internal, "%v", err)
	}
}

func toProtoItems(items []database.Item) ([]*proto.Item, error) {
	result := make([]*proto.Item, 0, len(items))
	for _, item := range items {
		converted, err := toProtoItem(item)
		if err != nil {
			return nil, err
		}
		result = append(result, converted)
	}

	return result, nil
}

func toProtoItem(item database.Item) (*proto.Item, error) {
	id, err := toProtoID(item.ID)
	if err != nil {
		return nil, err
	}

	result := &proto.Item{
		Id:          id,
		Title:       item.Title,
		Description: item.Description,
		Done:        item.Done,
		Position:    item.Position,
	}
	if item.DueDate != nil {
		result.DueDate = timestamppb.New(*item.DueDate)
	}
	if item.ListID != nil {
		listID, listErr := toProtoID(*item.ListID)
		if listErr != nil {
			return nil, listErr
		}
		result.ListId = &listID
	}

	labels, err := toProtoLabels(item.Labels)
	if err != nil {
		return nil, err
	}
	result.Labels = labels

	return result, nil
}

// toProtoID narrows a database identifier to the width used on the wire. The
// range check is not expected to trigger, but truncating an identifier would
// silently address the wrong record, so it is reported instead.
func toProtoID(id uint) (uint32, error) {
	if uint64(id) > math.MaxUint32 {
		return 0, status.Errorf(codes.Internal, "identifier %d exceeds the supported range", id)
	}

	return uint32(id), nil
}

// optionalID converts an optional protobuf identifier into the pointer form
// used by the database models. A zero identifier is treated as absent.
func optionalID(id *uint32) *uint {
	if id == nil || *id == 0 {
		return nil
	}
	value := uint(*id)

	return &value
}