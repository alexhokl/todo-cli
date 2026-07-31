package internal

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/alexhokl/todo-cli/database"
	"github.com/alexhokl/todo-cli/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// testUserID mirrors the identifier injected by the dummy authentication
// interceptor (uint(1)). The user is seeded by setupTodoServer so the per-user
// foreign keys resolve.
const testUserID uint = 1

// authenticatedContext returns a context carrying the test user identifier,
// matching what the dummy interceptor injects in production.
func authenticatedContext() context.Context {
	return context.WithValue(context.Background(), contextKeyUser{}, testUserID)
}

func setupTodoServer(t *testing.T) *TodoServer {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}
	if err := db.Create(&database.User{Username: "testuser"}).Error; err != nil {
		t.Fatalf("failed to seed the test user: %v", err)
	}

	return NewTodoServer(db)
}

// createTodos creates the named todos in order and returns their identifiers
// keyed by title.
func createTodos(t *testing.T, server *TodoServer, titles ...string) map[string]uint32 {
	t.Helper()

	ctx := authenticatedContext()
	ids := make(map[string]uint32, len(titles))
	for _, title := range titles {
		todo, err := server.CreateTodo(ctx, &proto.CreateTodoRequest{Title: title})
		if err != nil {
			t.Fatalf("failed to create the todo %q: %v", title, err)
		}
		ids[title] = todo.GetId()
	}

	return ids
}

func activeTitles(t *testing.T, server *TodoServer) []string {
	t.Helper()

	response, err := server.ListTodos(authenticatedContext(), &proto.ListTodosRequest{})
	if err != nil {
		t.Fatalf("failed to list the todos: %v", err)
	}

	titles := make([]string, 0, len(response.GetActive()))
	for _, todo := range response.GetActive() {
		titles = append(titles, todo.GetTitle())
	}

	return titles
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func TestCreateTodoAppendsToTail(t *testing.T) {
	server := setupTodoServer(t)
	createTodos(t, server, "a", "b", "c")

	if titles := activeTitles(t, server); !equalStrings(titles, []string{"a", "b", "c"}) {
		t.Errorf("expected [a b c] but got %v", titles)
	}
}

func TestCreateTodoRequiresTitle(t *testing.T) {
	server := setupTodoServer(t)

	_, err := server.CreateTodo(authenticatedContext(), &proto.CreateTodoRequest{})
	if got := status.Code(err); got != codes.InvalidArgument {
		t.Errorf("expected %v but got %v (%v)", codes.InvalidArgument, got, err)
	}
}

func TestMoveTodo(t *testing.T) {
	tests := []struct {
		name     string
		subject  string
		anchor   string
		before   bool
		expected []string
	}{
		{"before the head", "c", "a", true, []string{"c", "a", "b"}},
		{"after the tail", "a", "c", false, []string{"b", "c", "a"}},
		{"before the middle", "a", "b", true, []string{"a", "b", "c"}},
		{"after the middle", "a", "b", false, []string{"b", "a", "c"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := setupTodoServer(t)
			ids := createTodos(t, server, "a", "b", "c")

			req := &proto.MoveTodoRequest{Id: ids[test.subject]}
			if test.before {
				req.Anchor = &proto.MoveTodoRequest_BeforeId{BeforeId: ids[test.anchor]}
			} else {
				req.Anchor = &proto.MoveTodoRequest_AfterId{AfterId: ids[test.anchor]}
			}

			if _, err := server.MoveTodo(authenticatedContext(), req); err != nil {
				t.Fatalf("expected no error but got %v", err)
			}
			if titles := activeTitles(t, server); !equalStrings(titles, test.expected) {
				t.Errorf("expected %v but got %v", test.expected, titles)
			}
		})
	}
}

func TestMoveTodoErrorCodes(t *testing.T) {
	tests := []struct {
		name     string
		request  func(ids map[string]uint32) *proto.MoveTodoRequest
		complete string
		expected codes.Code
	}{
		{
			name: "missing id",
			request: func(ids map[string]uint32) *proto.MoveTodoRequest {
				return &proto.MoveTodoRequest{Anchor: &proto.MoveTodoRequest_BeforeId{BeforeId: ids["a"]}}
			},
			expected: codes.InvalidArgument,
		},
		{
			name: "missing anchor",
			request: func(ids map[string]uint32) *proto.MoveTodoRequest {
				return &proto.MoveTodoRequest{Id: ids["a"]}
			},
			expected: codes.InvalidArgument,
		},
		{
			name: "unknown subject",
			request: func(ids map[string]uint32) *proto.MoveTodoRequest {
				return &proto.MoveTodoRequest{
					Id:     404,
					Anchor: &proto.MoveTodoRequest_BeforeId{BeforeId: ids["a"]},
				}
			},
			expected: codes.NotFound,
		},
		{
			name: "unknown anchor",
			request: func(ids map[string]uint32) *proto.MoveTodoRequest {
				return &proto.MoveTodoRequest{
					Id:     ids["a"],
					Anchor: &proto.MoveTodoRequest_BeforeId{BeforeId: 404},
				}
			},
			expected: codes.NotFound,
		},
		{
			name: "completed subject",
			request: func(ids map[string]uint32) *proto.MoveTodoRequest {
				return &proto.MoveTodoRequest{
					Id:     ids["a"],
					Anchor: &proto.MoveTodoRequest_BeforeId{BeforeId: ids["b"]},
				}
			},
			complete: "a",
			expected: codes.FailedPrecondition,
		},
		{
			name: "completed anchor",
			request: func(ids map[string]uint32) *proto.MoveTodoRequest {
				return &proto.MoveTodoRequest{
					Id:     ids["a"],
					Anchor: &proto.MoveTodoRequest_BeforeId{BeforeId: ids["b"]},
				}
			},
			complete: "b",
			expected: codes.FailedPrecondition,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := setupTodoServer(t)
			ids := createTodos(t, server, "a", "b", "c")
			if test.complete != "" {
				req := &proto.SetTodoDoneRequest{Id: ids[test.complete], Done: true}
				if _, err := server.SetTodoDone(authenticatedContext(), req); err != nil {
					t.Fatalf("failed to complete %q: %v", test.complete, err)
				}
			}

			_, err := server.MoveTodo(authenticatedContext(), test.request(ids))
			if got := status.Code(err); got != test.expected {
				t.Errorf("expected %v but got %v (%v)", test.expected, got, err)
			}
		})
	}
}

func TestMoveTodoChangesList(t *testing.T) {
	server := setupTodoServer(t)
	ids := createTodos(t, server, "a", "b")

	list := database.List{Name: "work", UserID: testUserID}
	if err := server.DB.Create(&list).Error; err != nil {
		t.Fatalf("failed to create the list: %v", err)
	}
	listID := uint32(list.ID)

	moved, err := server.MoveTodo(authenticatedContext(), &proto.MoveTodoRequest{
		Id:         ids["b"],
		Anchor:     &proto.MoveTodoRequest_BeforeId{BeforeId: ids["a"]},
		ChangeList: true,
		ListId:     &listID,
	})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if moved.GetListId() != listID {
		t.Errorf("expected list %d but got %d", listID, moved.GetListId())
	}

	// change_list without a list identifier detaches the todo.
	moved, err = server.MoveTodo(authenticatedContext(), &proto.MoveTodoRequest{
		Id:         ids["b"],
		Anchor:     &proto.MoveTodoRequest_AfterId{AfterId: ids["a"]},
		ChangeList: true,
	})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if moved.ListId != nil {
		t.Errorf("expected the list to be cleared but got %d", moved.GetListId())
	}
}

func TestSetTodoDone(t *testing.T) {
	server := setupTodoServer(t)
	ids := createTodos(t, server, "a", "b", "c")

	completed, err := server.SetTodoDone(authenticatedContext(), &proto.SetTodoDoneRequest{Id: ids["b"], Done: true})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if !completed.GetDone() {
		t.Errorf("expected the todo to be completed")
	}
	if completed.Position != nil {
		t.Errorf("expected the position to be cleared but got %v", completed.GetPosition())
	}
	if titles := activeTitles(t, server); !equalStrings(titles, []string{"a", "c"}) {
		t.Errorf("expected [a c] but got %v", titles)
	}

	// Reopening appends to the tail rather than restoring the original slot.
	if _, err := server.SetTodoDone(authenticatedContext(), &proto.SetTodoDoneRequest{Id: ids["b"]}); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if titles := activeTitles(t, server); !equalStrings(titles, []string{"a", "c", "b"}) {
		t.Errorf("expected [a c b] but got %v", titles)
	}
}

func TestSetTodoDoneErrorCodes(t *testing.T) {
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
			server := setupTodoServer(t)
			_, err := server.SetTodoDone(authenticatedContext(), &proto.SetTodoDoneRequest{Id: test.id, Done: true})
			if got := status.Code(err); got != test.expected {
				t.Errorf("expected %v but got %v (%v)", test.expected, got, err)
			}
		})
	}
}

func TestListTodosSeparatesCompleted(t *testing.T) {
	server := setupTodoServer(t)
	ids := createTodos(t, server, "a", "b")

	if _, err := server.SetTodoDone(authenticatedContext(), &proto.SetTodoDoneRequest{Id: ids["a"], Done: true}); err != nil {
		t.Fatalf("failed to complete the todo: %v", err)
	}

	response, err := server.ListTodos(authenticatedContext(), &proto.ListTodosRequest{})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if len(response.GetActive()) != 1 || response.GetActive()[0].GetTitle() != "b" {
		t.Errorf("expected only [b] to be active but got %v", response.GetActive())
	}
	if len(response.GetCompleted()) != 1 || response.GetCompleted()[0].GetTitle() != "a" {
		t.Errorf("expected only [a] to be completed but got %v", response.GetCompleted())
	}
}

func TestUserIDFromContextMissing(t *testing.T) {
	_, err := userIDFromContext(context.Background())
	if got := status.Code(err); got != codes.Unauthenticated {
		t.Errorf("expected %v but got %v (%v)", codes.Unauthenticated, got, err)
	}
}

func TestListTodosRejectsUnauthenticated(t *testing.T) {
	server := setupTodoServer(t)
	_, err := server.ListTodos(context.Background(), &proto.ListTodosRequest{})
	if got := status.Code(err); got != codes.Unauthenticated {
		t.Errorf("expected %v but got %v (%v)", codes.Unauthenticated, got, err)
	}
}

func TestCreateTodoRejectsUnauthenticated(t *testing.T) {
	server := setupTodoServer(t)
	// The title check runs before the auth check, so a title is supplied to
	// reach the userIDFromContext branch.
	_, err := server.CreateTodo(context.Background(), &proto.CreateTodoRequest{Title: "a"})
	if got := status.Code(err); got != codes.Unauthenticated {
		t.Errorf("expected %v but got %v (%v)", codes.Unauthenticated, got, err)
	}
}

func TestMoveTodoRejectsUnauthenticated(t *testing.T) {
	server := setupTodoServer(t)
	// The id check runs before the auth check.
	_, err := server.MoveTodo(context.Background(), &proto.MoveTodoRequest{
		Id:     1,
		Anchor: &proto.MoveTodoRequest_BeforeId{BeforeId: 2},
	})
	if got := status.Code(err); got != codes.Unauthenticated {
		t.Errorf("expected %v but got %v (%v)", codes.Unauthenticated, got, err)
	}
}

func TestSetTodoDoneRejectsUnauthenticated(t *testing.T) {
	server := setupTodoServer(t)
	_, err := server.SetTodoDone(context.Background(), &proto.SetTodoDoneRequest{Id: 1, Done: true})
	if got := status.Code(err); got != codes.Unauthenticated {
		t.Errorf("expected %v but got %v (%v)", codes.Unauthenticated, got, err)
	}
}

func TestToProtoIDOutOfRange(t *testing.T) {
	overflow := uint(math.MaxUint32) + 1
	_, err := toProtoID(overflow)
	if got := status.Code(err); got != codes.Internal {
		t.Errorf("expected %v but got %v (%v)", codes.Internal, got, err)
	}
}

func TestToProtoTodoWithListAndDueDate(t *testing.T) {
	due := time.Date(2026, time.August, 15, 0, 0, 0, 0, time.UTC)
	listID := uint(3)
	todo := database.Todo{
		Title:   "a",
		DueDate: &due,
		ListID:  &listID,
	}
	todo.ID = 1

	result, err := toProtoTodo(todo)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if result.DueDate == nil {
		t.Fatalf("expected a due date but got nil")
	}
	if !result.GetDueDate().AsTime().Equal(due) {
		t.Errorf("expected due %v but got %v", due, result.GetDueDate().AsTime())
	}
	if result.ListId == nil || *result.ListId != 3 {
		t.Errorf("expected list id 3 but got %v", result.ListId)
	}
}

func TestMapDatabaseError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected codes.Code
	}{
		{"nil", nil, codes.OK},
		{"todo not found", database.ErrTodoNotFound, codes.NotFound},
		{"label not found", database.ErrLabelNotFound, codes.NotFound},
		{"label exists", database.ErrLabelExists, codes.AlreadyExists},
		{"label name empty", database.ErrLabelNameEmpty, codes.InvalidArgument},
		{"todo completed", database.ErrTodoCompleted, codes.FailedPrecondition},
		{"anchor completed", database.ErrAnchorCompleted, codes.FailedPrecondition},
		{"label in use", database.ErrLabelInUse, codes.FailedPrecondition},
		{"unknown error", errors.New("something broke"), codes.Internal},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mapped := mapDatabaseError(test.err)
			if test.err == nil {
				if mapped != nil {
					t.Errorf("expected nil but got %v", mapped)
				}
				return
			}
			if got := status.Code(mapped); got != test.expected {
				t.Errorf("expected %v but got %v (%v)", test.expected, got, mapped)
			}
		})
	}
}
