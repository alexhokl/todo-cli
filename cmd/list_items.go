package cmd

import (
	"fmt"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

type listItemsOptions struct {
	Labels        []string
	Untriaged     bool
	TimeSensitive bool
	Done          bool
	Search        string
}

var listItemsOpts listItemsOptions

// listItemsCmd lists the items, active ones in their manual order.
var listItemsCmd = &cobra.Command{
	Use:     "items",
	Aliases: []string{"item"},
	Short:   "List items in their manual order",
	Long: `List items.

By default only active, triaged items are listed, in their manual order.

Repeating --label narrows the result: only items carrying every one of the
given labels are shown.

At most one of --untriaged, --time-sensitive, and --done may be given. They
select a single bucket instead of the default active view.

  --untriaged      not done and carrying no manual ordering rank yet
  --time-sensitive not done and carrying a due date
  --done           completed items

--search narrows the result to items whose title or description contains the
given substring (case-insensitive).`,
	Example: `  todo list items
  todo list items --label urgent
  todo list items --label urgent --label work
  todo list items --untriaged
  todo list items --time-sensitive
  todo list items --search milk`,
	Args:        cobra.NoArgs,
	Annotations: map[string]string{annotationRequiresService: "true"},
	RunE:        runListItems,
}

func init() {
	listCmd.AddCommand(listItemsCmd)

	listItemsCmd.Flags().StringArrayVar(&listItemsOpts.Labels, "label", nil, "Only show items carrying this label (repeatable)")
	listItemsCmd.Flags().BoolVar(&listItemsOpts.Untriaged, "untriaged", false, "Only show active items without a manual ordering rank")
	listItemsCmd.Flags().BoolVar(&listItemsOpts.TimeSensitive, "time-sensitive", false, "Only show active items carrying a due date")
	listItemsCmd.Flags().BoolVar(&listItemsOpts.Done, "done", false, "Only show completed items")
	listItemsCmd.Flags().StringVar(&listItemsOpts.Search, "search", "", "Only show items whose title or description contains this substring (case-insensitive)")
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
		Search: listItemsOpts.Search,
	})
	if err != nil {
		return fmt.Errorf("failed to list the items: %w", err)
	}

	out := cmd.OutOrStdout()

	// The section header (e.g. "Items:", "Done:") is intentionally not
	// printed: the column header row from writeItemTable is the first line
	// of output for every view.
	return writeItemTable(out, response.GetActive())
}

// resolveItemView maps the mutually exclusive view flags to a proto ItemView.
// Setting none of them returns the default triaged active view.
func resolveItemView() (proto.ItemView, error) {
	set := 0
	if listItemsOpts.Untriaged {
		set++
	}
	if listItemsOpts.TimeSensitive {
		set++
	}
	if listItemsOpts.Done {
		set++
	}
	if set > 1 {
		return 0, fmt.Errorf("--untriaged, --time-sensitive, and --done are mutually exclusive")
	}

	switch {
	case listItemsOpts.Untriaged:
		return proto.ItemView_ITEM_VIEW_UNTRIAGED, nil
	case listItemsOpts.TimeSensitive:
		return proto.ItemView_ITEM_VIEW_TIME_SENSITIVE, nil
	case listItemsOpts.Done:
		return proto.ItemView_ITEM_VIEW_DONE, nil
	default:
		return proto.ItemView_ITEM_VIEW_TRIAGED, nil
	}
}
