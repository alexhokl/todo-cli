package cmd

import (
	"fmt"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

type updateItemOptions struct {
	AddLabels    []string
	RemoveLabels []string
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
names are matched case insensitively and stored in lower case.`,
		Example: `  todo update item 7 --add-label urgent
  todo update item 7 --add-label urgent --add-label work
  todo update item 7 --remove-label later
  todo update item 7 --add-label urgent --remove-label later`,
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{annotationRequiresService: "true"},
		RunE:        runUpdateItem,
	}

	flags := cmd.Flags()
	flags.StringArrayVar(&updateItemOpts.AddLabels, "add-label", nil, "Label to attach to the item (repeatable)")
	flags.StringArrayVar(&updateItemOpts.RemoveLabels, "remove-label", nil, "Label to detach from the item (repeatable)")

	// Without at least one of these the command would have nothing to do.
	cmd.MarkFlagsOneRequired("add-label", "remove-label")

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

	req := buildUpdateItemLabelsRequest(id, updateItemOpts)
	item, err := proto.NewItemServiceClient(conn).UpdateItemLabels(cmd.Context(), req)
	if err != nil {
		return fmt.Errorf("failed to update the item: %w", err)
	}

	return writeItemLine(cmd.OutOrStdout(), item)
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