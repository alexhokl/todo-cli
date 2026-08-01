package cmd

import (
	"fmt"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

type createBlockerOptions struct {
	Description string
}

var createBlockerOpts createBlockerOptions

// createBlockerCmd attaches a new blocker to an item.
var createBlockerCmd = &cobra.Command{
	Use:         "blocker [item-id]",
	Short:       "Create a blocker on an item",
	Example:     `  todo create blocker 7 --description "waiting on legal review"`,
	Args:        cobra.ExactArgs(1),
	Annotations: map[string]string{annotationRequiresService: "true"},
	RunE:        runCreateBlocker,
}

func init() {
	createCmd.AddCommand(createBlockerCmd)

	createBlockerCmd.Flags().StringVar(&createBlockerOpts.Description, "description", "", "Description of the blockage")
	_ = createBlockerCmd.MarkFlagRequired("description")
}

func runCreateBlocker(cmd *cobra.Command, args []string) error {
	itemID, err := parseID(args[0], "item")
	if err != nil {
		return err
	}

	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	blocker, err := proto.NewItemServiceClient(conn).CreateBlocker(
		cmd.Context(),
		&proto.CreateBlockerRequest{ItemId: itemID, Description: createBlockerOpts.Description},
	)
	if err != nil {
		return fmt.Errorf("failed to create the blocker: %w", err)
	}

	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "blocker %d on item %d\n", blocker.GetId(), itemID); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}