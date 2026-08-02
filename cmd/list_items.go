package cmd

import (
	"fmt"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

type listItemsOptions struct {
	Labels        []string
	Untriaged     bool
	Triaged       bool
	TimeSensitive bool
	Done          bool
}

var listItemsOpts listItemsOptions

// listItemsCmd lists the items, active ones in their manual order.
var listItemsCmd = &cobra.Command{
	Use:     "items",
	Aliases: []string{"item"},
	Short:   "List items in their manual order",
	Long: `List items.

Repeating --label narrows the result: only items carrying every one of the
given labels are shown.

At most one of --untriaged, --triaged, --time-sensitive, and --done may be
given. They select a single bucket instead of the default active-plus-completed
view.

  --untriaged      not done and carrying no manual ordering rank yet
  --triaged        not done and already placed in the manual order
  --time-sensitive not done and carrying a due date
  --done           completed items`,
	Example: `  todo list items
  todo list items --label urgent
  todo list items --label urgent --label work
  todo list items --untriaged
  todo list items --time-sensitive`,
	Args:        cobra.NoArgs,
	Annotations: map[string]string{annotationRequiresService: "true"},
	RunE:        runListItems,
}

func init() {
	listCmd.AddCommand(listItemsCmd)

	listItemsCmd.Flags().StringArrayVar(&listItemsOpts.Labels, "label", nil, "Only show items carrying this label (repeatable)")
	listItemsCmd.Flags().BoolVar(&listItemsOpts.Untriaged, "untriaged", false, "Only show active items without a manual ordering rank")
	listItemsCmd.Flags().BoolVar(&listItemsOpts.Triaged, "triaged", false, "Only show active items already placed in the manual order")
	listItemsCmd.Flags().BoolVar(&listItemsOpts.TimeSensitive, "time-sensitive", false, "Only show active items carrying a due date")
	listItemsCmd.Flags().BoolVar(&listItemsOpts.Done, "done", false, "Only show completed items")
}

func runListItems(cmd *cobra.Command, _ []string) error {
	view, err := resolveItemView()
	if err != nil {
		return err
	}

	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	response, err := proto.NewItemServiceClient(conn).ListItems(cmd.Context(), &proto.ListItemsRequest{
		Labels: listItemsOpts.Labels,
		View:   view,
	})
	if err != nil {
		return fmt.Errorf("failed to list the items: %w", err)
	}

	out := cmd.OutOrStdout()

	if view == proto.ItemView_ITEM_VIEW_UNSPECIFIED {
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

	header, numbered := viewHeader(view)
	if _, err := fmt.Fprintln(out, header); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}
	return writeItemTable(out, response.GetActive(), numbered)
}

// resolveItemView maps the mutually exclusive view flags to a proto ItemView.
// Setting none of them keeps the legacy two-bucket behaviour.
func resolveItemView() (proto.ItemView, error) {
	set := 0
	if listItemsOpts.Untriaged {
		set++
	}
	if listItemsOpts.Triaged {
		set++
	}
	if listItemsOpts.TimeSensitive {
		set++
	}
	if listItemsOpts.Done {
		set++
	}
	if set > 1 {
		return 0, fmt.Errorf("--untriaged, --triaged, --time-sensitive, and --done are mutually exclusive")
	}

	switch {
	case listItemsOpts.Untriaged:
		return proto.ItemView_ITEM_VIEW_UNTRIAGED, nil
	case listItemsOpts.Triaged:
		return proto.ItemView_ITEM_VIEW_TRIAGED, nil
	case listItemsOpts.TimeSensitive:
		return proto.ItemView_ITEM_VIEW_TIME_SENSITIVE, nil
	case listItemsOpts.Done:
		return proto.ItemView_ITEM_VIEW_DONE, nil
	default:
		return proto.ItemView_ITEM_VIEW_UNSPECIFIED, nil
	}
}

// viewHeader returns the section header and whether the table rows carry an
// ordinal for the selected view.
func viewHeader(view proto.ItemView) (string, bool) {
	switch view {
	case proto.ItemView_ITEM_VIEW_UNTRIAGED:
		return "Untriaged:", true
	case proto.ItemView_ITEM_VIEW_TRIAGED:
		return "Triaged:", true
	case proto.ItemView_ITEM_VIEW_TIME_SENSITIVE:
		return "Time-sensitive:", true
	case proto.ItemView_ITEM_VIEW_DONE:
		return "Done:", false
	default:
		return "Items:", true
	}
}
