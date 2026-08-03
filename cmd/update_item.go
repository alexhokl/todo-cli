package cmd

import (
	"fmt"
	"time"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type updateItemOptions struct {
	AddLabels    []string
	RemoveLabels []string
	AddLinks     []uint
	RemoveLinks  []uint
	DueDate      string
	ClearDueDate bool
	Title        string
	Description  string
}

var updateItemOpts updateItemOptions

// updateItemCmd patches an existing item. It currently covers labels only, and
// is the place further field updates should be added rather than growing more
// verb style subcommands under `update`.
var updateItemCmd = newUpdateItemCmd()

// newUpdateItemCmd builds the command with its flag rules applied. It is a
// constructor rather than a literal so that tests can obtain an instance with
// an unpolluted flag set.
func newUpdateItemCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "item [id]",
		Short: "Update an item",
		Long: `Update an item.

Labels named with --add-label are created automatically if they do not exist
yet. Labels named with --remove-label that are not known are ignored. Label
names are matched case insensitively and stored in lower case.

Items linked with --add-link must already exist; the link is symmetric, so
linking A to B also links B to A. Self-links and unknown or cross-user ids are
rejected. --remove-link detaches the link on both sides.

The title must be non-empty; an empty title is rejected. The description is
optional and stored verbatim; an empty string clears it.`,
		Example: `  todo update item 7 --add-label urgent
  todo update item 7 --add-label urgent --add-label work
  todo update item 7 --remove-label later
  todo update item 7 --add-label urgent --remove-label later
  todo update item 7 --add-link 3 --add-link 5
  todo update item 7 --remove-link 3
  todo update item 7 --due-date 2026-08-15
  todo update item 7 --clear-due-date
  todo update item 7 --title "new title"
  todo update item 7 --description "new details"
  todo update item 7 --title "renamed" --description ""`,
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{annotationRequiresService: "true"},
		RunE:        runUpdateItem,
	}

	flags := cmd.Flags()
	flags.StringArrayVar(&updateItemOpts.AddLabels, "add-label", nil, "Label to attach to the item (repeatable)")
	flags.StringArrayVar(&updateItemOpts.RemoveLabels, "remove-label", nil, "Label to detach from the item (repeatable)")
	flags.UintSliceVar(&updateItemOpts.AddLinks, "add-link", nil, "ID of an item to link to (repeatable, symmetric)")
	flags.UintSliceVar(&updateItemOpts.RemoveLinks, "remove-link", nil, "ID of an item to unlink (repeatable)")
	flags.StringVar(&updateItemOpts.DueDate, "due-date", "", "Due date of the item in YYYY-MM-DD format")
	flags.BoolVar(&updateItemOpts.ClearDueDate, "clear-due-date", false, "Remove the due date from the item")
	flags.StringVar(&updateItemOpts.Title, "title", "", "New title of the item")
	flags.StringVar(&updateItemOpts.Description, "description", "", "New description of the item (empty string clears it)")

	// Without at least one of these the command would have nothing to do.
	cmd.MarkFlagsOneRequired("add-label", "remove-label", "add-link", "remove-link", "due-date", "clear-due-date", "title", "description")
	cmd.MarkFlagsMutuallyExclusive("due-date", "clear-due-date")

	return cmd
}

func init() {
	updateCmd.AddCommand(updateItemCmd)
}

func runUpdateItem(cmd *cobra.Command, args []string) error {
	id, err := parseID(args[0], "item")
	if err != nil {
		return err
	}

	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	client := proto.NewItemServiceClient(conn)
	ctx := cmd.Context()
	var item *proto.Item

	// Title/description runs first so later mutations (labels, links) operate
	// on the renamed item and their returned item carries the new title.
	if cmd.Flags().Changed("title") || cmd.Flags().Changed("description") {
		req := &proto.UpdateItemRequest{Id: id}
		// The server requires a non-empty title and stores description
		// verbatim. When only one of the two flags is set, fetch the current
		// value of the other so the update leaves it unchanged.
		if !cmd.Flags().Changed("title") || !cmd.Flags().Changed("description") {
			current, err := client.GetItem(ctx, &proto.GetItemRequest{Id: id})
			if err != nil {
				return fmt.Errorf("failed to read the item before updating: %w", err)
			}
			if !cmd.Flags().Changed("title") {
				req.Title = current.GetTitle()
			}
			if !cmd.Flags().Changed("description") {
				req.Description = current.GetDescription()
			}
		}
		if cmd.Flags().Changed("title") {
			req.Title = updateItemOpts.Title
		}
		if cmd.Flags().Changed("description") {
			req.Description = updateItemOpts.Description
		}
		item, err = client.UpdateItem(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to update the item: %w", err)
		}
	}

	if cmd.Flags().Changed("due-date") {
		dueDate, err := parseDueDate(updateItemOpts.DueDate)
		if err != nil {
			return err
		}
		item, err = client.UpdateItemDueDate(ctx, &proto.UpdateItemDueDateRequest{Id: id, DueDate: timestamppb.New(dueDate)})
		if err != nil {
			return fmt.Errorf("failed to update the item due date: %w", err)
		}
	} else if cmd.Flags().Changed("clear-due-date") {
		item, err = client.UpdateItemDueDate(ctx, &proto.UpdateItemDueDateRequest{Id: id})
		if err != nil {
			return fmt.Errorf("failed to clear the item due date: %w", err)
		}
	}

	// Label and link mutations are separate RPCs. Labels run first so the
	// returned item (used for the final output) carries the fresh labels; if
	// links are also requested they run after and their result is printed.
	if len(updateItemOpts.AddLabels) > 0 || len(updateItemOpts.RemoveLabels) > 0 {
		item, err = client.UpdateItemLabels(ctx, buildUpdateItemLabelsRequest(id, updateItemOpts))
		if err != nil {
			return fmt.Errorf("failed to update the item labels: %w", err)
		}
	}

	if len(updateItemOpts.AddLinks) > 0 || len(updateItemOpts.RemoveLinks) > 0 {
		req, err := buildUpdateItemLinksRequest(id, updateItemOpts)
		if err != nil {
			return err
		}
		item, err = client.UpdateItemLinks(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to update the item links: %w", err)
		}
	}

	if item == nil {
		// Neither branch ran; MarkFlagsOneRequired prevents this in practice.
		return fmt.Errorf("nothing to update")
	}

	return writeItemLine(cmd.OutOrStdout(), item)
}

func parseDueDate(value string) (time.Time, error) {
	dueDate, err := time.ParseInLocation(time.DateOnly, value, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid due date %q: expected YYYY-MM-DD", value)
	}
	return dueDate, nil
}

// buildUpdateItemLabelsRequest assembles the wire request from the parsed
// add/remove label flags.
func buildUpdateItemLabelsRequest(id uint32, opts updateItemOptions) *proto.UpdateItemLabelsRequest {
	return &proto.UpdateItemLabelsRequest{
		Id:     id,
		Add:    opts.AddLabels,
		Remove: opts.RemoveLabels,
	}
}

// buildUpdateItemLinksRequest assembles the wire request from the parsed
// add/remove link flags.
func buildUpdateItemLinksRequest(id uint32, opts updateItemOptions) (*proto.UpdateItemLinksRequest, error) {
	add, err := toUint32Slice(opts.AddLinks)
	if err != nil {
		return nil, err
	}
	remove, err := toUint32Slice(opts.RemoveLinks)
	if err != nil {
		return nil, err
	}
	return &proto.UpdateItemLinksRequest{
		Id:     id,
		Add:    add,
		Remove: remove,
	}, nil
}
