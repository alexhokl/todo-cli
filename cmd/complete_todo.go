package cmd

import (
	"fmt"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

type completeTodoOptions struct {
	Undo bool
}

var completeTodoOpts completeTodoOptions

// completeTodoCmd completes or reopens a todo.
var completeTodoCmd = &cobra.Command{
	Use:   "done [id]",
	Short: "Complete or reopen a todo",
	Long: `Complete a todo, which removes it from the manual ordering, or reopen
one with --undo, which appends it to the end of the manual ordering.`,
	Example: `  todo update done 7
  todo update done 7 --undo`,
	Args:        cobra.ExactArgs(1),
	Annotations: map[string]string{annotationRequiresService: "true"},
	RunE:        runCompleteTodo,
}

func init() {
	updateCmd.AddCommand(completeTodoCmd)

	completeTodoCmd.Flags().BoolVar(&completeTodoOpts.Undo, "undo", false, "Reopen the todo instead of completing it")
}

func runCompleteTodo(cmd *cobra.Command, args []string) error {
	id, err := parseID(args[0], "todo")
	if err != nil {
		return err
	}

	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	req := buildSetTodoDoneRequest(id, completeTodoOpts.Undo)
	todo, err := proto.NewTodoServiceClient(conn).SetTodoDone(cmd.Context(), req)
	if err != nil {
		return fmt.Errorf("failed to update the todo: %w", err)
	}

	return writeTodoLine(cmd.OutOrStdout(), todo)
}

// buildSetTodoDoneRequest assembles the wire request. Without --undo the todo
// is completed; with --undo it is reopened.
func buildSetTodoDoneRequest(id uint32, undo bool) *proto.SetTodoDoneRequest {
	return &proto.SetTodoDoneRequest{Id: id, Done: !undo}
}
