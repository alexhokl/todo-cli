package cmd

import (
	"fmt"
	"strconv"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

type moveItemOptions struct {
	BeforeID  uint32
	AfterID   uint32
	Top       bool
	Bottom    bool
	ListID    uint32
	ClearList bool
}

var moveItemOpts moveItemOptions

// moveItemCmd reprioritises an item within the manual ordering.
var moveItemCmd = newMoveItemCmd()

// newMoveItemCmd builds the command with its flag rules applied. It is a
// constructor rather than a literal so that tests can obtain an instance with
// an unpolluted flag set.
func newMoveItemCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "priority [id]",
		Short: "Prioritise an item relative to another item or at an end of the order",
		Long: `Prioritise an item to a new place in the manual ordering.

The new place is expressed relative to another triaged item, so the ordering
stays stable no matter how many items are added or completed in between.
Untriaged items (including newly created ones) can be triaged with --top or
--bottom, which is also the only way to prioritise the first item.`,
		Example: `  todo update priority 7 --before 3
  todo update priority 7 --after 3
  todo update priority 7 --top
  todo update priority 7 --bottom
  todo update priority 7 --before 3 --list 2
  todo update priority 7 --after 3 --clear-list`,
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{annotationRequiresService: "true"},
		RunE:        runMoveItem,
	}

	flags := cmd.Flags()
	flags.Uint32Var(&moveItemOpts.BeforeID, "before", 0, "ID of the item to move this item before (anchor must be triaged)")
	flags.Uint32Var(&moveItemOpts.AfterID, "after", 0, "ID of the item to move this item after (anchor must be triaged)")
	flags.BoolVar(&moveItemOpts.Top, "top", false, "Assign the highest priority (triage an untriaged item)")
	flags.BoolVar(&moveItemOpts.Bottom, "bottom", false, "Assign the lowest priority (triage an untriaged item)")
	flags.Uint32Var(&moveItemOpts.ListID, "list", 0, "ID of the list to move this item to")
	flags.BoolVar(&moveItemOpts.ClearList, "clear-list", false, "Detach this item from its list")

	// Exactly one anchor is required and single valued, and an item cannot be
	// both assigned to a list and detached from one.
	cmd.MarkFlagsMutuallyExclusive("before", "after")
	cmd.MarkFlagsMutuallyExclusive("before", "top")
	cmd.MarkFlagsMutuallyExclusive("before", "bottom")
	cmd.MarkFlagsMutuallyExclusive("after", "top")
	cmd.MarkFlagsMutuallyExclusive("after", "bottom")
	cmd.MarkFlagsMutuallyExclusive("top", "bottom")
	cmd.MarkFlagsOneRequired("before", "after", "top", "bottom")
	cmd.MarkFlagsMutuallyExclusive("list", "clear-list")

	return cmd
}

func init() {
	updateCmd.AddCommand(moveItemCmd)
}

func runMoveItem(cmd *cobra.Command, args []string) error {
	id, err := parseID(args[0], "item")
	if err != nil {
		return err
	}

	req := buildMoveItemRequest(id, moveItemOpts, cmd.Flags().Changed("list"))

	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	item, err := proto.NewItemServiceClient(conn).MoveItem(cmd.Context(), req)
	if err != nil {
		return fmt.Errorf("failed to move the item: %w", err)
	}

	return writeItemLine(cmd.OutOrStdout(), item)
}

// buildMoveItemRequest assembles the wire request from the parsed flags. The
// anchor is selected from before/after/top/bottom, and change_list
// distinguishes leaving the list alone from clearing it: it is set by either
// --clear-list or --list, while the identifier is only sent for --list.
func buildMoveItemRequest(id uint32, opts moveItemOptions, listChanged bool) *proto.MoveItemRequest {
	req := &proto.MoveItemRequest{Id: id}
	switch {
	case opts.BeforeID != 0:
		req.Anchor = &proto.MoveItemRequest_BeforeId{BeforeId: opts.BeforeID}
	case opts.AfterID != 0:
		req.Anchor = &proto.MoveItemRequest_AfterId{AfterId: opts.AfterID}
	case opts.Top:
		req.Anchor = &proto.MoveItemRequest_Top{Top: true}
	case opts.Bottom:
		req.Anchor = &proto.MoveItemRequest_Bottom{Bottom: true}
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
// a label ID where an item ID is expected produces a message the user can act
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