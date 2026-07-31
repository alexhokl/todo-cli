package cmd

import (
	"fmt"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

type listItemsOptions struct {
	Labels []string
}

var listItemsOpts listItemsOptions

// listItemsCmd lists the items, active ones in their manual order.
var listItemsCmd = &cobra.Command{
	Use:   "items",
	Short: "List items in their manual order",
	Long: `List items.

Repeating --label narrows the result: only items carrying every one of the
given labels are shown.`,
	Example: `  todo list items
  todo list items --label urgent
  todo list items --label urgent --label work`,
	Args:        cobra.NoArgs,
	Annotations: map[string]string{annotationRequiresService: "true"},
	RunE:        runListItems,
}

func init() {
	listCmd.AddCommand(listItemsCmd)

	listItemsCmd.Flags().StringArrayVar(&listItemsOpts.Labels, "label", nil, "Only show items carrying this label (repeatable)")
}

func runListItems(cmd *cobra.Command, _ []string) error {
	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	response, err := proto.NewItemServiceClient(conn).ListItems(cmd.Context(), &proto.ListItemsRequest{Labels: listItemsOpts.Labels})
	if err != nil {
		return fmt.Errorf("failed to list the items: %w", err)
	}

	out := cmd.OutOrStdout()
	if _, err := fmt.Fprintln(out, "Active:"); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}
	if err := writeItemTable(out, response.GetActive(), true); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(out, "\nCompleted:"); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	// Completed items are not part of the manual ordering, so they are listed
	// by how recently they were updated and carry no ordinal.
	return writeItemTable(out, response.GetCompleted(), false)
}