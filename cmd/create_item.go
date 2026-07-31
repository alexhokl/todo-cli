package cmd

import (
	"fmt"
	"time"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type createItemOptions struct {
	Description string
	DueDate     string
	ListID      uint32
	Labels      []string
}

var createItemOpts createItemOptions

// createItemCmd creates an item at the end of the manual ordering.
var createItemCmd = &cobra.Command{
	Use:   "item [title]",
	Short: "Create an item at the end of the manual order",
	Example: `  todo create item "buy milk" --due 2026-08-15
  todo create item "ship it" --label urgent --label work`,
	Args:        cobra.ExactArgs(1),
	Annotations: map[string]string{annotationRequiresService: "true"},
	RunE:        runCreateItem,
}

func init() {
	createCmd.AddCommand(createItemCmd)

	flags := createItemCmd.Flags()
	flags.StringVarP(&createItemOpts.Description, "description", "d", "", "Description of the item")
	flags.StringVar(&createItemOpts.DueDate, "due", "", "Due date of the item in YYYY-MM-DD format")
	flags.Uint32Var(&createItemOpts.ListID, "list", 0, "ID of the list to add this item to")
	flags.StringArrayVar(&createItemOpts.Labels, "label", nil, "Label to attach to the item, created if unknown (repeatable)")
}

func runCreateItem(cmd *cobra.Command, args []string) error {
	req, err := buildCreateItemRequest(args, createItemOpts, cmd.Flags().Changed("list"))
	if err != nil {
		return err
	}

	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	item, err := proto.NewItemServiceClient(conn).CreateItem(cmd.Context(), req)
	if err != nil {
		return fmt.Errorf("failed to create the item: %w", err)
	}

	return writeItemLine(cmd.OutOrStdout(), item)
}

// buildCreateItemRequest assembles the wire request from the parsed flags. The
// due date is interpreted in the local time zone so a date entered by the user
// means that day where they are, not in UTC. A zero list identifier is left
// absent unless the flag was explicitly set.
func buildCreateItemRequest(args []string, opts createItemOptions, listChanged bool) (*proto.CreateItemRequest, error) {
	req := &proto.CreateItemRequest{
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