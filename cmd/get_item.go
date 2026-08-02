package cmd

import (
	"fmt"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

// getItemCmd retrieves a single item by identifier and prints its details.
var getItemCmd = &cobra.Command{
	Use:         "item [id]",
	Short:       "Get an item",
	Args:        cobra.ExactArgs(1),
	Annotations: map[string]string{annotationRequiresService: "true"},
	RunE:        runGetItem,
}

func init() {
	getCmd.AddCommand(getItemCmd)
}

func runGetItem(cmd *cobra.Command, args []string) error {
	id, err := parseID(args[0], "item")
	if err != nil {
		return err
	}

	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	item, err := proto.NewItemServiceClient(conn).GetItem(cmd.Context(), &proto.GetItemRequest{Id: id})
	if err != nil {
		return fmt.Errorf("failed to get the item: %w", err)
	}

	return writeItemDetail(cmd.OutOrStdout(), item)
}