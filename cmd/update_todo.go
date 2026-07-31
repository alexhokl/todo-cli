package cmd

import (
	"fmt"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

type updateTodoOptions struct {
	AddLabels    []string
	RemoveLabels []string
}

var updateTodoOpts updateTodoOptions

// updateTodoCmd patches an existing todo. It currently covers labels only, and
// is the place further field updates should be added rather than growing more
// verb style subcommands under `update`.
var updateTodoCmd = newUpdateTodoCmd()

// newUpdateTodoCmd builds the command with its flag rules applied. It is a
// constructor rather than a literal so that tests can obtain an instance with
// an unpolluted flag set.
func newUpdateTodoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "todo [id]",
		Short: "Update a todo",
		Long: `Update a todo.

Labels named with --add-label are created automatically if they do not exist
yet. Labels named with --remove-label that are not known are ignored. Label
names are matched case insensitively and stored in lower case.`,
		Example: `  todo update todo 7 --add-label urgent
  todo update todo 7 --add-label urgent --add-label work
  todo update todo 7 --remove-label later
  todo update todo 7 --add-label urgent --remove-label later`,
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{annotationRequiresService: "true"},
		RunE:        runUpdateTodo,
	}

	flags := cmd.Flags()
	flags.StringArrayVar(&updateTodoOpts.AddLabels, "add-label", nil, "Label to attach to the todo (repeatable)")
	flags.StringArrayVar(&updateTodoOpts.RemoveLabels, "remove-label", nil, "Label to detach from the todo (repeatable)")

	// Without at least one of these the command would have nothing to do.
	cmd.MarkFlagsOneRequired("add-label", "remove-label")

	return cmd
}

func init() {
	updateCmd.AddCommand(updateTodoCmd)
}

func runUpdateTodo(cmd *cobra.Command, args []string) error {
	id, err := parseID(args[0], "todo")
	if err != nil {
		return err
	}

	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	req := buildUpdateTodoLabelsRequest(id, updateTodoOpts)
	todo, err := proto.NewTodoServiceClient(conn).UpdateTodoLabels(cmd.Context(), req)
	if err != nil {
		return fmt.Errorf("failed to update the todo: %w", err)
	}

	return writeTodoLine(cmd.OutOrStdout(), todo)
}

// buildUpdateTodoLabelsRequest assembles the wire request from the parsed
// add/remove label flags.
func buildUpdateTodoLabelsRequest(id uint32, opts updateTodoOptions) *proto.UpdateTodoLabelsRequest {
	return &proto.UpdateTodoLabelsRequest{
		Id:     id,
		Add:    opts.AddLabels,
		Remove: opts.RemoveLabels,
	}
}
