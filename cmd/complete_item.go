package cmd

import (
	"fmt"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

type completeItemOptions struct {
	Undo bool
}

var completeItemOpts completeItemOptions

// completeItemCmd completes or reopens an item.
var completeItemCmd = &cobra.Command{
	Use:   "done [id]",
	Short: "Complete or reopen an item",
	Long: `Complete an item, which removes it from the manual ordering, or reopen
one with --undo, which appends it to the end of the manual ordering.`,
	Example: `  todo update done 7
  todo update done 7 --undo`,
	Args:        cobra.ExactArgs(1),
	Annotations: map[string]string{annotationRequiresService: "true"},
	RunE:        runCompleteItem,
}

func init() {
	updateCmd.AddCommand(completeItemCmd)

	completeItemCmd.Flags().BoolVar(&completeItemOpts.Undo, "undo", false, "Reopen the item instead of completing it")
}

func runCompleteItem(cmd *cobra.Command, args []string) error {
	id, err := parseID(args[0], "item")
	if err != nil {
		return err
	}

	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	req := buildSetItemDoneRequest(id, completeItemOpts.Undo)
	item, err := proto.NewItemServiceClient(conn).SetItemDone(cmd.Context(), req)
	if err != nil {
		return fmt.Errorf("failed to update the item: %w", err)
	}

	return writeItemLine(cmd.OutOrStdout(), item)
}

// buildSetItemDoneRequest assembles the wire request. Without --undo the item
// is completed; with --undo it is reopened.
func buildSetItemDoneRequest(id uint32, undo bool) *proto.SetItemDoneRequest {
	return &proto.SetItemDoneRequest{Id: id, Done: !undo}
}