package cmd

import (
	"testing"
	"time"

	"github.com/alexhokl/todo-cli/proto"
)

func TestBuildCreateTodoRequest(t *testing.T) {
	t.Run("title only", func(t *testing.T) {
		req, err := buildCreateTodoRequest([]string{"buy milk"}, createTodoOptions{}, false)
		if err != nil {
			t.Fatalf("expected no error but got %v", err)
		}
		if req.GetTitle() != "buy milk" {
			t.Errorf("expected title %q but got %q", "buy milk", req.GetTitle())
		}
		if req.DueDate != nil {
			t.Errorf("expected no due date but got %v", req.GetDueDate())
		}
		if req.ListId != nil {
			t.Errorf("expected no list id but got %v", req.ListId)
		}
	})

	t.Run("with description and labels", func(t *testing.T) {
		req, err := buildCreateTodoRequest(
			[]string{"ship it"},
			createTodoOptions{Description: "before friday", Labels: []string{"urgent", "work"}},
			false,
		)
		if err != nil {
			t.Fatalf("expected no error but got %v", err)
		}
		if req.GetDescription() != "before friday" {
			t.Errorf("expected description %q but got %q", "before friday", req.GetDescription())
		}
		if len(req.GetLabels()) != 2 {
			t.Errorf("expected 2 labels but got %d", len(req.GetLabels()))
		}
	})

	t.Run("with valid due date", func(t *testing.T) {
		req, err := buildCreateTodoRequest(
			[]string{"a"},
			createTodoOptions{DueDate: "2026-08-15"},
			false,
		)
		if err != nil {
			t.Fatalf("expected no error but got %v", err)
		}
		if req.DueDate == nil {
			t.Fatalf("expected a due date but got nil")
		}
		expected := time.Date(2026, time.August, 15, 0, 0, 0, 0, time.Local)
		if !req.GetDueDate().AsTime().Equal(expected) {
			t.Errorf("expected %v but got %v", expected, req.GetDueDate().AsTime())
		}
	})

	t.Run("with invalid due date", func(t *testing.T) {
		_, err := buildCreateTodoRequest(
			[]string{"a"},
			createTodoOptions{DueDate: "15-08-2026"},
			false,
		)
		if err == nil {
			t.Fatalf("expected an error but got none")
		}
	})

	t.Run("list set explicitly", func(t *testing.T) {
		req, err := buildCreateTodoRequest(
			[]string{"a"},
			createTodoOptions{ListID: 5},
			true,
		)
		if err != nil {
			t.Fatalf("expected no error but got %v", err)
		}
		if req.ListId == nil || *req.ListId != 5 {
			t.Errorf("expected list id 5 but got %v", req.ListId)
		}
	})

	t.Run("list flag unchanged leaves list absent", func(t *testing.T) {
		req, err := buildCreateTodoRequest(
			[]string{"a"},
			createTodoOptions{ListID: 5},
			false,
		)
		if err != nil {
			t.Fatalf("expected no error but got %v", err)
		}
		if req.ListId != nil {
			t.Errorf("expected no list id but got %v", req.ListId)
		}
	})
}

func TestBuildMoveTodoRequest(t *testing.T) {
	tests := []struct {
		name        string
		id          uint32
		opts        moveTodoOptions
		listChanged bool
		check       func(*testing.T, *proto.MoveTodoRequest)
	}{
		{
			name: "before anchor",
			id:   7,
			opts: moveTodoOptions{BeforeID: 3},
			check: func(t *testing.T, r *proto.MoveTodoRequest) {
				before, ok := r.GetAnchor().(*proto.MoveTodoRequest_BeforeId)
				if !ok {
					t.Fatalf("expected a BeforeId anchor but got %T", r.GetAnchor())
				}
				if before.BeforeId != 3 {
					t.Errorf("expected before id 3 but got %d", before.BeforeId)
				}
			},
		},
		{
			name: "after anchor",
			id:   7,
			opts: moveTodoOptions{AfterID: 4},
			check: func(t *testing.T, r *proto.MoveTodoRequest) {
				after, ok := r.GetAnchor().(*proto.MoveTodoRequest_AfterId)
				if !ok {
					t.Fatalf("expected an AfterId anchor but got %T", r.GetAnchor())
				}
				if after.AfterId != 4 {
					t.Errorf("expected after id 4 but got %d", after.AfterId)
				}
			},
		},
		{
			name: "clear list",
			id:   7,
			opts: moveTodoOptions{BeforeID: 3, ClearList: true},
			check: func(t *testing.T, r *proto.MoveTodoRequest) {
				if !r.GetChangeList() {
					t.Errorf("expected change_list to be set")
				}
				if r.ListId != nil {
					t.Errorf("expected no list id on clear but got %v", r.ListId)
				}
			},
		},
		{
			name:        "set list",
			id:          7,
			opts:        moveTodoOptions{BeforeID: 3, ListID: 2},
			listChanged: true,
			check: func(t *testing.T, r *proto.MoveTodoRequest) {
				if !r.GetChangeList() {
					t.Errorf("expected change_list to be set")
				}
				if r.ListId == nil || *r.ListId != 2 {
					t.Errorf("expected list id 2 but got %v", r.ListId)
				}
			},
		},
		{
			name:        "neither clear nor list leaves list untouched",
			id:          7,
			opts:        moveTodoOptions{BeforeID: 3, ListID: 2},
			listChanged: false,
			check: func(t *testing.T, r *proto.MoveTodoRequest) {
				if r.GetChangeList() {
					t.Errorf("expected change_list to be unset")
				}
				if r.ListId != nil {
					t.Errorf("expected no list id but got %v", r.ListId)
				}
			},
		},
		{
			name: "clear list takes precedence over list flag",
			id:   7,
			opts: moveTodoOptions{BeforeID: 3, ClearList: true, ListID: 2},
			// listChanged is true but ClearList wins, so no identifier is sent.
			listChanged: true,
			check: func(t *testing.T, r *proto.MoveTodoRequest) {
				if !r.GetChangeList() {
					t.Errorf("expected change_list to be set")
				}
				if r.ListId != nil {
					t.Errorf("expected no list id on clear but got %v", r.ListId)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := buildMoveTodoRequest(test.id, test.opts, test.listChanged)
			if req.GetId() != test.id {
				t.Errorf("expected id %d but got %d", test.id, req.GetId())
			}
			test.check(t, req)
		})
	}
}

func TestBuildSetTodoDoneRequest(t *testing.T) {
	t.Run("complete", func(t *testing.T) {
		req := buildSetTodoDoneRequest(7, false)
		if req.GetId() != 7 {
			t.Errorf("expected id 7 but got %d", req.GetId())
		}
		if !req.GetDone() {
			t.Errorf("expected done to be true")
		}
	})

	t.Run("undo", func(t *testing.T) {
		req := buildSetTodoDoneRequest(7, true)
		if req.GetDone() {
			t.Errorf("expected done to be false")
		}
	})
}

func TestBuildUpdateTodoLabelsRequest(t *testing.T) {
	req := buildUpdateTodoLabelsRequest(7, updateTodoOptions{
		AddLabels:    []string{"urgent", "work"},
		RemoveLabels: []string{"later"},
	})
	if req.GetId() != 7 {
		t.Errorf("expected id 7 but got %d", req.GetId())
	}
	if len(req.GetAdd()) != 2 || req.GetAdd()[0] != "urgent" || req.GetAdd()[1] != "work" {
		t.Errorf("expected add [urgent work] but got %v", req.GetAdd())
	}
	if len(req.GetRemove()) != 1 || req.GetRemove()[0] != "later" {
		t.Errorf("expected remove [later] but got %v", req.GetRemove())
	}
}
