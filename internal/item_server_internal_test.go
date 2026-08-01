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
// interceptor (uint(1)). The user is seeded by setupItemServer so the per-user
// foreign keys resolve.
const testUserID uint = 1

// authenticatedContext returns a context carrying the test user identifier,
// matching what the dummy interceptor injects in production.
func authenticatedContext() context.Context {
	return context.WithValue(context.Background(), contextKeyUser{}, testUserID)
}

func setupItemServer(t *testing.T) *ItemServer {
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

	return NewItemServer(db)
}

// createItems creates the named items in order, triages each via MoveItem with
// bottom (so the first item is highest priority and the last is lowest), and
// returns their identifiers keyed by title. The display order therefore
// matches the creation order.
func createItems(t *testing.T, server *ItemServer, titles ...string) map[string]uint32 {
	t.Helper()

	ctx := authenticatedContext()
	ids := make(map[string]uint32, len(titles))
	for _, title := range titles {
		item, err := server.CreateItem(ctx, &proto.CreateItemRequest{Title: title})
		if err != nil {
			t.Fatalf("failed to create the item %q: %v", title, err)
		}
		ids[title] = item.GetId()
	}

	// Triage in creation order by appending to the tail (bottom). The first
	// item lands at priority 0 on an empty ordering, the next at -priorityStep,
	// and so on, so the display order (DESC) matches the creation order.
	for _, title := range titles {
		req := &proto.MoveItemRequest{Id: ids[title], Anchor: &proto.MoveItemRequest_Bottom{Bottom: true}}
		if _, err := server.MoveItem(ctx, req); err != nil {
			t.Fatalf("failed to triage %q: %v", title, err)
		}
	}

	return ids
}

func activeTitles(t *testing.T, server *ItemServer) []string {
	t.Helper()

	response, err := server.ListItems(authenticatedContext(), &proto.ListItemsRequest{})
	if err != nil {
		t.Fatalf("failed to list the items: %v", err)
	}

	titles := make([]string, 0, len(response.GetActive()))
	for _, item := range response.GetActive() {
		titles = append(titles, item.GetTitle())
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

func TestCreateItemLeavesItemUntriaged(t *testing.T) {
	server := setupItemServer(t)

	// Create items directly via CreateItem (no triage) so they remain
	// untriaged, which is the new default.
	ctx := authenticatedContext()
	for _, title := range []string{"a", "b", "c"} {
		if _, err := server.CreateItem(ctx, &proto.CreateItemRequest{Title: title}); err != nil {
			t.Fatalf("failed to create %q: %v", title, err)
		}
	}
	// Newly created items are untriaged: they carry no priority and are
	// excluded from the default active listing.
	response, err := server.ListItems(ctx, &proto.ListItemsRequest{})
	if err != nil {
		t.Fatalf("failed to list the items: %v", err)
	}
	if len(response.GetActive()) != 0 {
		t.Errorf("expected no triaged active items but got %v", response.GetActive())
	}

	// The items show up in the untriaged view instead, in creation order.
	untriaged, err := server.ListItems(ctx, &proto.ListItemsRequest{View: proto.ItemView_ITEM_VIEW_UNTRIAGED})
	if err != nil {
		t.Fatalf("failed to list the untriaged items: %v", err)
	}
	titles := make([]string, 0, len(untriaged.GetActive()))
	for _, item := range untriaged.GetActive() {
		titles = append(titles, item.GetTitle())
	}
	if !equalStrings(titles, []string{"a", "b", "c"}) {
		t.Errorf("expected [a b c] untriaged but got %v", titles)
	}
}

func TestCreateItemRequiresTitle(t *testing.T) {
	server := setupItemServer(t)

	_, err := server.CreateItem(authenticatedContext(), &proto.CreateItemRequest{})
	if got := status.Code(err); got != codes.InvalidArgument {
		t.Errorf("expected %v but got %v (%v)", codes.InvalidArgument, got, err)
	}
}

func TestCreateItemWithEffort(t *testing.T) {
	server := setupItemServer(t)
	if _, err := server.CreateEffort(authenticatedContext(), &proto.CreateEffortRequest{Name: "high"}); err != nil {
		t.Fatalf("failed to create the effort: %v", err)
	}

	high := "High"
	item, err := server.CreateItem(authenticatedContext(), &proto.CreateItemRequest{
		Title:  "task",
		Effort: &high,
	})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if item.GetEffort() == nil || item.GetEffort().GetName() != "high" {
		t.Errorf("expected effort high but got %v", item.GetEffort())
	}

	// An unknown effort name is reported rather than being created.
	unknown := "unknown"
	unknownReq := &proto.CreateItemRequest{
		Title:  "task2",
		Effort: &unknown,
	}
	_, err = server.CreateItem(authenticatedContext(), unknownReq)
	if got := status.Code(err); got != codes.NotFound {
		t.Errorf("expected %v but got %v (%v)", codes.NotFound, got, err)
	}
}

func TestMoveItem(t *testing.T) {
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
			server := setupItemServer(t)
			ids := createItems(t, server, "a", "b", "c")

			req := &proto.MoveItemRequest{Id: ids[test.subject]}
			if test.before {
				req.Anchor = &proto.MoveItemRequest_BeforeId{BeforeId: ids[test.anchor]}
			} else {
				req.Anchor = &proto.MoveItemRequest_AfterId{AfterId: ids[test.anchor]}
			}

			if _, err := server.MoveItem(authenticatedContext(), req); err != nil {
				t.Fatalf("expected no error but got %v", err)
			}
			if titles := activeTitles(t, server); !equalStrings(titles, test.expected) {
				t.Errorf("expected %v but got %v", test.expected, titles)
			}
		})
	}
}

func TestMoveItemErrorCodes(t *testing.T) {
	tests := []struct {
		name           string
		request        func(ids map[string]uint32) *proto.MoveItemRequest
		complete       string
		createUntriaged bool
		expected       codes.Code
	}{
		{
			name: "missing id",
			request: func(ids map[string]uint32) *proto.MoveItemRequest {
				return &proto.MoveItemRequest{Anchor: &proto.MoveItemRequest_BeforeId{BeforeId: ids["a"]}}
			},
			expected: codes.InvalidArgument,
		},
		{
			name: "missing anchor",
			request: func(ids map[string]uint32) *proto.MoveItemRequest {
				return &proto.MoveItemRequest{Id: ids["a"]}
			},
			expected: codes.InvalidArgument,
		},
		{
			name: "unknown subject",
			request: func(ids map[string]uint32) *proto.MoveItemRequest {
				return &proto.MoveItemRequest{
					Id:     404,
					Anchor: &proto.MoveItemRequest_BeforeId{BeforeId: ids["a"]},
				}
			},
			expected: codes.NotFound,
		},
		{
			name: "unknown anchor",
			request: func(ids map[string]uint32) *proto.MoveItemRequest {
				return &proto.MoveItemRequest{
					Id:     ids["a"],
					Anchor: &proto.MoveItemRequest_BeforeId{BeforeId: 404},
				}
			},
			expected: codes.NotFound,
		},
		{
			name: "completed subject",
			request: func(ids map[string]uint32) *proto.MoveItemRequest {
				return &proto.MoveItemRequest{
					Id:     ids["a"],
					Anchor: &proto.MoveItemRequest_BeforeId{BeforeId: ids["b"]},
				}
			},
			complete: "a",
			expected: codes.FailedPrecondition,
		},
		{
			name: "completed anchor",
			request: func(ids map[string]uint32) *proto.MoveItemRequest {
				return &proto.MoveItemRequest{
					Id:     ids["a"],
					Anchor: &proto.MoveItemRequest_BeforeId{BeforeId: ids["b"]},
				}
			},
			complete: "b",
			expected: codes.FailedPrecondition,
		},
		{
			name: "untriaged anchor",
			request: func(ids map[string]uint32) *proto.MoveItemRequest {
				// "d" is created untriaged and never triaged, so moving
				// relative to it is rejected with InvalidArgument.
				return &proto.MoveItemRequest{
					Id:     ids["a"],
					Anchor: &proto.MoveItemRequest_BeforeId{BeforeId: ids["d"]},
				}
			},
			createUntriaged: true,
			expected:       codes.InvalidArgument,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := setupItemServer(t)
			ids := createItems(t, server, "a", "b", "c")
			if test.createUntriaged {
				// "d" is created directly (no triage) so it carries no priority.
				more := createItems(t, server, "d")
				for k, v := range more {
					ids[k] = v
				}
				// Undo the triage that createItems applied, leaving d untriaged.
				if err := server.DB.Model(&database.Item{}).
					Where("id = ?", ids["d"]).
					Update("priority", nil).Error; err != nil {
					t.Fatalf("failed to clear the priority of d: %v", err)
				}
			}
			if test.complete != "" {
				req := &proto.SetItemDoneRequest{Id: ids[test.complete], Done: true}
				if _, err := server.SetItemDone(authenticatedContext(), req); err != nil {
					t.Fatalf("failed to complete %q: %v", test.complete, err)
				}
			}

			_, err := server.MoveItem(authenticatedContext(), test.request(ids))
			if got := status.Code(err); got != test.expected {
				t.Errorf("expected %v but got %v (%v)", test.expected, got, err)
			}
		})
	}
}

func TestMoveItemChangesList(t *testing.T) {
	server := setupItemServer(t)
	ids := createItems(t, server, "a", "b")

	list := database.List{Name: "work", UserID: testUserID}
	if err := server.DB.Create(&list).Error; err != nil {
		t.Fatalf("failed to create the list: %v", err)
	}
	listID := uint32(list.ID)

	moved, err := server.MoveItem(authenticatedContext(), &proto.MoveItemRequest{
		Id:         ids["b"],
		Anchor:     &proto.MoveItemRequest_BeforeId{BeforeId: ids["a"]},
		ChangeList: true,
		ListId:     &listID,
	})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if moved.GetListId() != listID {
		t.Errorf("expected list %d but got %d", listID, moved.GetListId())
	}

	// change_list without a list identifier detaches the item.
	moved, err = server.MoveItem(authenticatedContext(), &proto.MoveItemRequest{
		Id:         ids["b"],
		Anchor:     &proto.MoveItemRequest_AfterId{AfterId: ids["a"]},
		ChangeList: true,
	})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if moved.ListId != nil {
		t.Errorf("expected the list to be cleared but got %d", moved.GetListId())
	}
}

func TestSetItemDone(t *testing.T) {
	server := setupItemServer(t)
	ids := createItems(t, server, "a", "b", "c")

	completed, err := server.SetItemDone(authenticatedContext(), &proto.SetItemDoneRequest{Id: ids["b"], Done: true})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if !completed.GetDone() {
		t.Errorf("expected the item to be completed")
	}
	if completed.Priority != nil {
		t.Errorf("expected the priority to be cleared but got %v", completed.GetPriority())
	}
	if titles := activeTitles(t, server); !equalStrings(titles, []string{"a", "c"}) {
		t.Errorf("expected [a c] but got %v", titles)
	}

	// Reopening returns the item to the untriaged bucket rather than appending
	// it to the manual order, so it carries no priority and does not appear in
	// the default active listing.
	reopened, err := server.SetItemDone(authenticatedContext(), &proto.SetItemDoneRequest{Id: ids["b"]})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if reopened.Priority != nil {
		t.Fatalf("expected no priority after reopening but got %v", reopened.GetPriority())
	}
	if titles := activeTitles(t, server); !equalStrings(titles, []string{"a", "c"}) {
		t.Errorf("expected [a c] with the reopened item untriaged but got %v", titles)
	}
	// The reopened item shows up in the untriaged view instead.
	untriaged, err := server.ListItems(authenticatedContext(), &proto.ListItemsRequest{View: proto.ItemView_ITEM_VIEW_UNTRIAGED})
	if err != nil {
		t.Fatalf("failed to list the untriaged items: %v", err)
	}
	untriagedTitles := make([]string, 0, len(untriaged.GetActive()))
	for _, item := range untriaged.GetActive() {
		untriagedTitles = append(untriagedTitles, item.GetTitle())
	}
	if !equalStrings(untriagedTitles, []string{"b"}) {
		t.Errorf("expected [b] untriaged but got %v", untriagedTitles)
	}
}

func TestSetItemDoneErrorCodes(t *testing.T) {
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
			server := setupItemServer(t)
			_, err := server.SetItemDone(authenticatedContext(), &proto.SetItemDoneRequest{Id: test.id, Done: true})
			if got := status.Code(err); got != test.expected {
				t.Errorf("expected %v but got %v (%v)", test.expected, got, err)
			}
		})
	}
}

func TestListItemsSeparatesCompleted(t *testing.T) {
	server := setupItemServer(t)
	ids := createItems(t, server, "a", "b")

	if _, err := server.SetItemDone(authenticatedContext(), &proto.SetItemDoneRequest{Id: ids["a"], Done: true}); err != nil {
		t.Fatalf("failed to complete the item: %v", err)
	}

	response, err := server.ListItems(authenticatedContext(), &proto.ListItemsRequest{})
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

func TestListItemsByView(t *testing.T) {
	server := setupItemServer(t)
	ids := createItems(t, server, "triaged-a", "triaged-b", "triaged-c")

	// Create an untriaged item directly so it never receives a priority.
	untriaged := database.Item{Title: "untriaged-d", UserID: testUserID}
	if err := server.DB.Create(&untriaged).Error; err != nil {
		t.Fatalf("failed to create the untriaged item: %v", err)
	}

	// Give triaged-c a due date so it is also time-sensitive.
	due := time.Date(2026, time.August, 15, 0, 0, 0, 0, time.UTC)
	if err := server.DB.Model(&database.Item{}).Where("id = ?", ids["triaged-c"]).
		UpdateColumn("due_date", due).Error; err != nil {
		t.Fatalf("failed to set the due date: %v", err)
	}

	// Complete triaged-b.
	if _, err := server.SetItemDone(authenticatedContext(), &proto.SetItemDoneRequest{Id: ids["triaged-b"], Done: true}); err != nil {
		t.Fatalf("failed to complete the item: %v", err)
	}

	tests := []struct {
		name     string
		view     proto.ItemView
		expected []string
	}{
		{"untriaged", proto.ItemView_ITEM_VIEW_UNTRIAGED, []string{"untriaged-d"}},
		{"triaged", proto.ItemView_ITEM_VIEW_TRIAGED, []string{"triaged-a", "triaged-c"}},
		{"time-sensitive", proto.ItemView_ITEM_VIEW_TIME_SENSITIVE, []string{"triaged-c"}},
		{"done", proto.ItemView_ITEM_VIEW_DONE, []string{"triaged-b"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, err := server.ListItems(authenticatedContext(), &proto.ListItemsRequest{View: test.view})
			if err != nil {
				t.Fatalf("expected no error but got %v", err)
			}
			if len(response.GetCompleted()) != 0 {
				t.Errorf("expected no completed items but got %v", response.GetCompleted())
			}

			titles := make([]string, 0, len(response.GetActive()))
			for _, item := range response.GetActive() {
				titles = append(titles, item.GetTitle())
			}
			if !equalStrings(titles, test.expected) {
				t.Errorf("expected %v but got %v", test.expected, titles)
			}
		})
	}
}

func TestListItemsUnspecifiedKeepsBothBuckets(t *testing.T) {
	server := setupItemServer(t)
	ids := createItems(t, server, "a", "b")

	if _, err := server.SetItemDone(authenticatedContext(), &proto.SetItemDoneRequest{Id: ids["a"], Done: true}); err != nil {
		t.Fatalf("failed to complete the item: %v", err)
	}

	response, err := server.ListItems(authenticatedContext(), &proto.ListItemsRequest{})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if len(response.GetActive()) != 1 || response.GetActive()[0].GetTitle() != "b" {
		t.Errorf("expected [b] active but got %v", response.GetActive())
	}
	if len(response.GetCompleted()) != 1 || response.GetCompleted()[0].GetTitle() != "a" {
		t.Errorf("expected [a] completed but got %v", response.GetCompleted())
	}
}

func TestUserIDFromContextMissing(t *testing.T) {
	_, err := userIDFromContext(context.Background())
	if got := status.Code(err); got != codes.Unauthenticated {
		t.Errorf("expected %v but got %v (%v)", codes.Unauthenticated, got, err)
	}
}

func TestListItemsRejectsUnauthenticated(t *testing.T) {
	server := setupItemServer(t)
	_, err := server.ListItems(context.Background(), &proto.ListItemsRequest{})
	if got := status.Code(err); got != codes.Unauthenticated {
		t.Errorf("expected %v but got %v (%v)", codes.Unauthenticated, got, err)
	}
}

func TestCreateItemRejectsUnauthenticated(t *testing.T) {
	server := setupItemServer(t)
	// The title check runs before the auth check, so a title is supplied to
	// reach the userIDFromContext branch.
	_, err := server.CreateItem(context.Background(), &proto.CreateItemRequest{Title: "a"})
	if got := status.Code(err); got != codes.Unauthenticated {
		t.Errorf("expected %v but got %v (%v)", codes.Unauthenticated, got, err)
	}
}

func TestMoveItemRejectsUnauthenticated(t *testing.T) {
	server := setupItemServer(t)
	// The id check runs before the auth check.
	_, err := server.MoveItem(context.Background(), &proto.MoveItemRequest{
		Id:     1,
		Anchor: &proto.MoveItemRequest_BeforeId{BeforeId: 2},
	})
	if got := status.Code(err); got != codes.Unauthenticated {
		t.Errorf("expected %v but got %v (%v)", codes.Unauthenticated, got, err)
	}
}

func TestSetItemDoneRejectsUnauthenticated(t *testing.T) {
	server := setupItemServer(t)
	_, err := server.SetItemDone(context.Background(), &proto.SetItemDoneRequest{Id: 1, Done: true})
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

func TestToProtoItemWithListAndDueDate(t *testing.T) {
	due := time.Date(2026, time.August, 15, 0, 0, 0, 0, time.UTC)
	listID := uint(3)
	item := database.Item{
		Title:   "a",
		DueDate: &due,
		ListID:  &listID,
	}
	item.ID = 1

	result, err := toProtoItem(item)
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
		{"item not found", database.ErrItemNotFound, codes.NotFound},
		{"label not found", database.ErrLabelNotFound, codes.NotFound},
		{"label exists", database.ErrLabelExists, codes.AlreadyExists},
		{"label name empty", database.ErrLabelNameEmpty, codes.InvalidArgument},
		{"item completed", database.ErrItemCompleted, codes.FailedPrecondition},
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