package cmd

import (
	"fmt"
	"strconv"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

type moveTodoOptions struct {
	BeforeID  uint32
	AfterID   uint32
	ListID    uint32
	ClearList bool
}

var moveTodoOpts moveTodoOptions

// moveTodoCmd repositions a todo within the manual ordering.
var moveTodoCmd = newMoveTodoCmd()

// newMoveTodoCmd builds the command with its flag rules applied. It is a
// constructor rather than a literal so that tests can obtain an instance with
// an unpolluted flag set.
func newMoveTodoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "position [id]",
		Short: "Move a todo before or after another todo",
		Long: `Move a todo to a new place in the manual ordering.

The new place is expressed relative to another todo, so the ordering stays
stable no matter how many todos are added or completed in between.`,
		Example: `  todo update position 7 --before 3
  todo update position 7 --after 3
  todo update position 7 --before 3 --list 2
  todo update position 7 --after 3 --clear-list`,
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{annotationRequiresService: "true"},
		RunE:        runMoveTodo,
	}

	flags := cmd.Flags()
	flags.Uint32Var(&moveTodoOpts.BeforeID, "before", 0, "ID of the todo to move this todo before")
	flags.Uint32Var(&moveTodoOpts.AfterID, "after", 0, "ID of the todo to move this todo after")
	flags.Uint32Var(&moveTodoOpts.ListID, "list", 0, "ID of the list to move this todo to")
	flags.BoolVar(&moveTodoOpts.ClearList, "clear-list", false, "Detach this todo from its list")

	// The anchor is required and single valued, and a todo cannot be both
	// assigned to a list and detached from one.
	cmd.MarkFlagsMutuallyExclusive("before", "after")
	cmd.MarkFlagsOneRequired("before", "after")
	cmd.MarkFlagsMutuallyExclusive("list", "clear-list")

	return cmd
}

func init() {
	updateCmd.AddCommand(moveTodoCmd)
}

func runMoveTodo(cmd *cobra.Command, args []string) error {
	id, err := parseID(args[0], "todo")
	if err != nil {
		return err
	}

	req := buildMoveTodoRequest(id, moveTodoOpts, cmd.Flags().Changed("list"))

	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	todo, err := proto.NewTodoServiceClient(conn).MoveTodo(cmd.Context(), req)
	if err != nil {
		return fmt.Errorf("failed to move the todo: %w", err)
	}

	return writeTodoLine(cmd.OutOrStdout(), todo)
}

// buildMoveTodoRequest assembles the wire request from the parsed flags. The
// anchor is selected from before/after, and change_list distinguishes leaving
// the list alone from clearing it: it is set by either --clear-list or --list,
// while the identifier is only sent for --list.
func buildMoveTodoRequest(id uint32, opts moveTodoOptions, listChanged bool) *proto.MoveTodoRequest {
	req := &proto.MoveTodoRequest{Id: id}
	if opts.BeforeID != 0 {
		req.Anchor = &proto.MoveTodoRequest_BeforeId{BeforeId: opts.BeforeID}
	} else {
		req.Anchor = &proto.MoveTodoRequest_AfterId{AfterId: opts.AfterID}
	}

	if opts.ClearList {
		req.ChangeList = true
	} else if listChanged {
		req.ChangeList = true
		req.ListId = &opts.ListID
	}

	return req
}

// parseID converts a positional identifier argument into the unsigned form used
// on the wire. The subject names what the identifier refers to so that passing
// a label ID where a todo ID is expected produces a message the user can act
// on.
func parseID(value, subject string) (uint32, error) {
	id, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid %s ID %q: %w", subject, value, err)
	}
	if id == 0 {
		return 0, fmt.Errorf("invalid %s ID %q: must be greater than zero", subject, value)
	}

	return uint32(id), nil
}
