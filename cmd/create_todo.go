package cmd

import (
	"fmt"
	"time"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type createTodoOptions struct {
	Description string
	DueDate     string
	ListID      uint32
	Labels      []string
}

var createTodoOpts createTodoOptions

// createTodoCmd creates a todo at the end of the manual ordering.
var createTodoCmd = &cobra.Command{
	Use:   "todo [title]",
	Short: "Create a todo at the end of the manual order",
	Example: `  todo create todo "buy milk" --due 2026-08-15
  todo create todo "ship it" --label urgent --label work`,
	Args:        cobra.ExactArgs(1),
	Annotations: map[string]string{annotationRequiresService: "true"},
	RunE:        runCreateTodo,
}

func init() {
	createCmd.AddCommand(createTodoCmd)

	flags := createTodoCmd.Flags()
	flags.StringVarP(&createTodoOpts.Description, "description", "d", "", "Description of the todo")
	flags.StringVar(&createTodoOpts.DueDate, "due", "", "Due date of the todo in YYYY-MM-DD format")
	flags.Uint32Var(&createTodoOpts.ListID, "list", 0, "ID of the list to add this todo to")
	flags.StringArrayVar(&createTodoOpts.Labels, "label", nil, "Label to attach to the todo, created if unknown (repeatable)")
}

func runCreateTodo(cmd *cobra.Command, args []string) error {
	req, err := buildCreateTodoRequest(args, createTodoOpts, cmd.Flags().Changed("list"))
	if err != nil {
		return err
	}

	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	todo, err := proto.NewTodoServiceClient(conn).CreateTodo(cmd.Context(), req)
	if err != nil {
		return fmt.Errorf("failed to create the todo: %w", err)
	}

	return writeTodoLine(cmd.OutOrStdout(), todo)
}

// buildCreateTodoRequest assembles the wire request from the parsed flags. The
// due date is interpreted in the local time zone so a date entered by the user
// means that day where they are, not in UTC. A zero list identifier is left
// absent unless the flag was explicitly set.
func buildCreateTodoRequest(args []string, opts createTodoOptions, listChanged bool) (*proto.CreateTodoRequest, error) {
	req := &proto.CreateTodoRequest{
		Title:       args[0],
		Description: opts.Description,
		Labels:      opts.Labels,
	}

	if opts.DueDate != "" {
		dueDate, err := time.ParseInLocation(time.DateOnly, opts.DueDate, time.Local)
		if err != nil {
			return nil, fmt.Errorf("invalid due date %q: expected YYYY-MM-DD", opts.DueDate)
		}
		req.DueDate = timestamppb.New(dueDate)
	}

	if listChanged {
		req.ListId = &opts.ListID
	}

	return req, nil
}
