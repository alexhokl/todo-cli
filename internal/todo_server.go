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

// TodoServer implements the todo gRPC service on top of a GORM database.
type TodoServer struct {
	proto.UnimplementedTodoServiceServer
	DB *gorm.DB
}

// NewTodoServer creates a todo service implementation backed by the given
// database connection.
func NewTodoServer(db *gorm.DB) *TodoServer {
	return &TodoServer{DB: db}
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

// ListTodos returns the active todos in manual order and the completed todos
// ordered by how recently they were updated.
func (s *TodoServer) ListTodos(ctx context.Context, req *proto.ListTodosRequest) (*proto.ListTodosResponse, error) {
	ctx, span := startSpan(ctx, "ListTodos")
	defer span.End()

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	filter := database.TodoFilter{Labels: req.GetLabels()}

	active, err := database.ListActive(s.DB.WithContext(ctx), userID, filter)
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	completed, err := database.ListCompleted(s.DB.WithContext(ctx), userID, filter)
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	activeTodos, err := toProtoTodos(active)
	if err != nil {
		return nil, err
	}
	completedTodos, err := toProtoTodos(completed)
	if err != nil {
		return nil, err
	}

	endSpanOk(span)

	return &proto.ListTodosResponse{
		Active:    activeTodos,
		Completed: completedTodos,
	}, nil
}

// CreateTodo appends a new todo to the end of the manual order.
func (s *TodoServer) CreateTodo(ctx context.Context, req *proto.CreateTodoRequest) (*proto.Todo, error) {
	ctx, span := startSpan(ctx, "CreateTodo")
	defer span.End()

	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	todo := database.Todo{
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
		ListID:      optionalID(req.ListId),
	}
	if req.DueDate != nil {
		dueDate := req.GetDueDate().AsTime()
		todo.DueDate = &dueDate
	}

	db := s.DB.WithContext(ctx)
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := database.AssignInitialPosition(tx, &todo, userID); err != nil {
			return err
		}
		// Labels are resolved before the insert so that the todo is created
		// with its join rows already in place.
		labels, err := database.FindOrCreateLabels(tx, userID, req.GetLabels())
		if err != nil {
			return err
		}
		todo.Labels = labels

		return tx.Create(&todo).Error
	})
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	result, err := toProtoTodo(todo)
	if err != nil {
		return nil, err
	}

	endSpanOk(span)

	return result, nil
}

// MoveTodo places a todo immediately before or after another todo, optionally
// reassigning its list in the same operation.
func (s *TodoServer) MoveTodo(ctx context.Context, req *proto.MoveTodoRequest) (*proto.Todo, error) {
	ctx, span := startSpan(ctx, "MoveTodo")
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

	todo, err := database.MoveTodo(s.DB.WithContext(ctx), userID, uint(req.GetId()), anchor, opts)
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

// SetTodoDone completes or reopens a todo.
func (s *TodoServer) SetTodoDone(ctx context.Context, req *proto.SetTodoDoneRequest) (*proto.Todo, error) {
	ctx, span := startSpan(ctx, "SetTodoDone")
	defer span.End()

	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	todo, err := database.SetDone(s.DB.WithContext(ctx), userID, uint(req.GetId()), req.GetDone())
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

// mapDatabaseError translates the sentinel errors of the database package into
// gRPC status errors. The error logging interceptor takes care of recording
// them, so handlers only have to return the result.
func mapDatabaseError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, database.ErrTodoNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, database.ErrLabelNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, database.ErrLabelExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, database.ErrLabelNameEmpty):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, database.ErrTodoCompleted),
		errors.Is(err, database.ErrAnchorCompleted),
		errors.Is(err, database.ErrLabelInUse):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Errorf(codes.Internal, "%v", err)
	}
}

func toProtoTodos(todos []database.Todo) ([]*proto.Todo, error) {
	result := make([]*proto.Todo, 0, len(todos))
	for _, todo := range todos {
		converted, err := toProtoTodo(todo)
		if err != nil {
			return nil, err
		}
		result = append(result, converted)
	}

	return result, nil
}

func toProtoTodo(todo database.Todo) (*proto.Todo, error) {
	id, err := toProtoID(todo.ID)
	if err != nil {
		return nil, err
	}

	result := &proto.Todo{
		Id:          id,
		Title:       todo.Title,
		Description: todo.Description,
		Done:        todo.Done,
		Position:    todo.Position,
	}
	if todo.DueDate != nil {
		result.DueDate = timestamppb.New(*todo.DueDate)
	}
	if todo.ListID != nil {
		listID, listErr := toProtoID(*todo.ListID)
		if listErr != nil {
			return nil, listErr
		}
		result.ListId = &listID
	}

	labels, err := toProtoLabels(todo.Labels)
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
