package cmd

import (
	"fmt"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

// deleteItemCmd deletes an untriaged item.
var deleteItemCmd = &cobra.Command{
	Use:   "item [id]",
	Short: "Delete an untriaged item",
	Long: `Delete an untriaged item.

Only items that are not done and carry no priority (i.e. are still untriaged)
may be deleted. An item that has linked items must be unlinked first.
Attached blockers and comments are removed in the same operation.`,
	Example:     `  todo delete item 3`,
	Args:        cobra.ExactArgs(1),
	Annotations: map[string]string{annotationRequiresService: "true"},
	RunE:        runDeleteItem,
}

func init() {
	deleteCmd.AddCommand(deleteItemCmd)
}

func runDeleteItem(cmd *cobra.Command, args []string) error {
	id, err := parseID(args[0], "item")
	if err != nil {
		return err
	}

	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	if _, err := proto.NewItemServiceClient(conn).DeleteItem(
		cmd.Context(),
		&proto.DeleteItemRequest{Id: id},
	); err != nil {
		return fmt.Errorf("failed to delete the item: %w", err)
	}

	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "deleted item %d\n", id); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}