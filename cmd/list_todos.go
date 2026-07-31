package cmd

import (
	"fmt"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

type listTodosOptions struct {
	Labels []string
}

var listTodosOpts listTodosOptions

// listTodosCmd lists the todos, active ones in their manual order.
var listTodosCmd = &cobra.Command{
	Use:   "todos",
	Short: "List todos in their manual order",
	Long: `List todos.

Repeating --label narrows the result: only todos carrying every one of the
given labels are shown.`,
	Example: `  todo list todos
  todo list todos --label urgent
  todo list todos --label urgent --label work`,
	Args:        cobra.NoArgs,
	Annotations: map[string]string{annotationRequiresService: "true"},
	RunE:        runListTodos,
}

func init() {
	listCmd.AddCommand(listTodosCmd)

	listTodosCmd.Flags().StringArrayVar(&listTodosOpts.Labels, "label", nil, "Only show todos carrying this label (repeatable)")
}

func runListTodos(cmd *cobra.Command, _ []string) error {
	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	response, err := proto.NewTodoServiceClient(conn).ListTodos(cmd.Context(), &proto.ListTodosRequest{Labels: listTodosOpts.Labels})
	if err != nil {
		return fmt.Errorf("failed to list the todos: %w", err)
	}

	out := cmd.OutOrStdout()
	if _, err := fmt.Fprintln(out, "Active:"); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}
	if err := writeTodoTable(out, response.GetActive(), true); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(out, "\nCompleted:"); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	// Completed todos are not part of the manual ordering, so they are listed
	// by how recently they were updated and carry no ordinal.
	return writeTodoTable(out, response.GetCompleted(), false)
}
