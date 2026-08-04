package internal

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/alexhokl/todo-cli/database"
	"github.com/alexhokl/todo-cli/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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

	filter := database.ItemFilter{
		Labels: req.GetLabels(),
		View:   mapItemView(req.GetView()),
		Search: req.GetSearch(),
	}

	if filter.View == database.ItemViewUnspecified {
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

	items, err := database.ListItemsByView(s.DB.WithContext(ctx), userID, filter)
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	activeItems, err := toProtoItems(items)
	if err != nil {
		return nil, err
	}

	endSpanOk(span)

	return &proto.ListItemsResponse{Active: activeItems}, nil
}

// mapItemView translates the proto enum into the database-layer enum so the
// database package stays free of generated-code imports.
func mapItemView(view proto.ItemView) database.ItemView {
	switch view {
	case proto.ItemView_ITEM_VIEW_UNTRIAGED:
		return database.ItemViewUntriaged
	case proto.ItemView_ITEM_VIEW_TRIAGED:
		return database.ItemViewTriaged
	case proto.ItemView_ITEM_VIEW_TIME_SENSITIVE:
		return database.ItemViewTimeSensitive
	case proto.ItemView_ITEM_VIEW_DONE:
		return database.ItemViewDone
	default:
		return database.ItemViewUnspecified
	}
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
		UserID:      userID,
	}
	if req.DueDate != nil {
		dueDate := req.GetDueDate().AsTime()
		item.DueDate = &dueDate
	}

	db := s.DB.WithContext(ctx)
	err = db.Transaction(func(tx *gorm.DB) error {
		// Items are created untriaged (Priority is nil). Triage happens via
		// MoveItem with top/bottom or a relative anchor, so there is no
		// initial priority to assign here.
		// Labels are resolved before the insert so that the item is created
		// with its join rows already in place.
		labels, err := database.FindOrCreateLabels(tx, userID, req.GetLabels())
		if err != nil {
			return err
		}
		item.Labels = labels

		// An effort name, if supplied, must resolve to an existing effort;
		// unlike labels it is not created on the fly so the effort catalog
		// stays an explicit step.
		if req.GetEffort() != "" {
			effort, err := database.FindEffortByName(tx, userID, req.GetEffort())
			if err != nil {
				return err
			}
			item.EffortID = &effort.ID
			// Set the association so the returned item carries the effort
			// without a second fetch (labels are populated the same way).
			item.Effort = effort
		}

		if err := tx.Create(&item).Error; err != nil {
			return err
		}

		// Linked items are attached after the insert so the item has an id and
		// the symmetric two-row join writes can reference it. The target ids
		// are resolved (and scoped to the caller) inside UpdateItemLinks, which
		// also reloads the item with the preloaded associations.
		if len(req.GetLinkItemIds()) > 0 {
			ids := make([]uint, 0, len(req.GetLinkItemIds()))
			for _, id := range req.GetLinkItemIds() {
				ids = append(ids, uint(id))
			}
			linked, err := database.UpdateItemLinks(tx, userID, item.ID, ids, nil)
			if err != nil {
				return err
			}
			item = *linked
		}

		return nil
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

// GetItem returns a single item by identifier.
func (s *ItemServer) GetItem(ctx context.Context, req *proto.GetItemRequest) (*proto.Item, error) {
	ctx, span := startSpan(ctx, "GetItem")
	defer span.End()

	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	item, err := database.GetItem(s.DB.WithContext(ctx), userID, uint(req.GetId()))
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

// DeleteItem removes an untriaged item. Only items that are not done and carry
// no priority may be deleted; items with linked items are rejected. Attached
// blockers and comments are removed in the same operation.
func (s *ItemServer) DeleteItem(ctx context.Context, req *proto.DeleteItemRequest) (*emptypb.Empty, error) {
	ctx, span := startSpan(ctx, "DeleteItem")
	defer span.End()

	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := database.DeleteItem(s.DB.WithContext(ctx), userID, uint(req.GetId())); err != nil {
		return nil, mapDatabaseError(err)
	}

	endSpanOk(span)
	return &emptypb.Empty{}, nil
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
	switch req.GetAnchor().(type) {
	case *proto.MoveItemRequest_BeforeId:
		anchor = database.MoveAnchor{TargetID: uint(req.GetBeforeId()), Before: true}
	case *proto.MoveItemRequest_AfterId:
		anchor = database.MoveAnchor{TargetID: uint(req.GetAfterId()), Before: false}
	case *proto.MoveItemRequest_Top:
		anchor = database.MoveAnchor{Top: true}
	case *proto.MoveItemRequest_Bottom:
		anchor = database.MoveAnchor{Bottom: true}
	default:
		return nil, status.Error(codes.InvalidArgument, "exactly one of before_id, after_id, top, or bottom is required")
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

// UpdateItemDueDate sets or clears an item's due date.
func (s *ItemServer) UpdateItemDueDate(ctx context.Context, req *proto.UpdateItemDueDateRequest) (*proto.Item, error) {
	ctx, span := startSpan(ctx, "UpdateItemDueDate")
	defer span.End()

	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var dueDate *time.Time
	if req.GetDueDate() != nil {
		value := req.GetDueDate().AsTime()
		dueDate = &value
	}

	item, err := database.UpdateItemDueDate(s.DB.WithContext(ctx), userID, uint(req.GetId()), dueDate)
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

// UpdateItem changes an item's title and description. The title must be
// non-empty after trimming; an empty description clears the field.
func (s *ItemServer) UpdateItem(ctx context.Context, req *proto.UpdateItemRequest) (*proto.Item, error) {
	ctx, span := startSpan(ctx, "UpdateItem")
	defer span.End()

	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	item, err := database.UpdateItem(s.DB.WithContext(ctx), userID, uint(req.GetId()), req.GetTitle(), req.GetDescription())
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
	case errors.Is(err, database.ErrLabelNotFound),
		errors.Is(err, database.ErrEffortNotFound),
		errors.Is(err, database.ErrBlockerNotFound),
		errors.Is(err, database.ErrCommentNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, database.ErrLabelExists),
		errors.Is(err, database.ErrEffortExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, database.ErrLabelNameEmpty),
		errors.Is(err, database.ErrLabelColourInvalid),
		errors.Is(err, database.ErrEffortNameEmpty),
		errors.Is(err, database.ErrBlockerDescriptionEmpty),
		errors.Is(err, database.ErrCommentBodyEmpty),
		errors.Is(err, database.ErrItemTitleEmpty),
		errors.Is(err, database.ErrItemLinkToSelf):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, database.ErrItemCompleted),
		errors.Is(err, database.ErrAnchorCompleted),
		errors.Is(err, database.ErrItemNotUntriaged),
		errors.Is(err, database.ErrItemHasLinks),
		errors.Is(err, database.ErrLabelInUse),
		errors.Is(err, database.ErrEffortInUse):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, database.ErrAnchorUntriaged):
		return status.Error(codes.InvalidArgument, err.Error())
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
		Priority:    item.Priority,
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

	if item.Effort != nil {
		result.Effort, err = toProtoEffort(*item.Effort)
		if err != nil {
			return nil, err
		}
	}

	blockers, err := toProtoBlockers(item.Blockers)
	if err != nil {
		return nil, err
	}
	result.Blockers = blockers

	comments, err := toProtoComments(item.Comments)
	if err != nil {
		return nil, err
	}
	result.Comments = comments

	linkedItems, err := toProtoItems(item.LinkedItems)
	if err != nil {
		return nil, err
	}
	result.LinkedItems = linkedItems

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
