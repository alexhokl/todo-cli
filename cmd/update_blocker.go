package cmd

import (
	"fmt"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

type updateBlockerOptions struct {
	Description string
}

var updateBlockerOpts updateBlockerOptions

// updateBlockerCmd changes the description of an existing blocker.
var updateBlockerCmd = &cobra.Command{
	Use:         "blocker [blocker-id]",
	Short:       "Update a blocker description",
	Example:     `  todo update blocker 3 --description "still waiting on legal"`,
	Args:        cobra.ExactArgs(1),
	Annotations: map[string]string{annotationRequiresService: "true"},
	RunE:        runUpdateBlocker,
}

func init() {
	updateCmd.AddCommand(updateBlockerCmd)

	updateBlockerCmd.Flags().StringVar(&updateBlockerOpts.Description, "description", "", "New description of the blockage")
	_ = updateBlockerCmd.MarkFlagRequired("description")
}

func runUpdateBlocker(cmd *cobra.Command, args []string) error {
	id, err := parseID(args[0], "blocker")
	if err != nil {
		return err
	}

	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	blocker, err := proto.NewItemServiceClient(conn).UpdateBlocker(
		cmd.Context(),
		&proto.UpdateBlockerRequest{Id: id, Description: updateBlockerOpts.Description},
	)
	if err != nil {
		return fmt.Errorf("failed to update the blocker: %w", err)
	}

	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "blocker %d updated\n", blocker.GetId()); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}