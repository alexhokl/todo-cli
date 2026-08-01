package cmd

import (
	"fmt"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

// deleteBlockerCmd deletes a blocker.
var deleteBlockerCmd = &cobra.Command{
	Use:         "blocker [blocker-id]",
	Short:       "Delete a blocker",
	Example:     `  todo delete blocker 3`,
	Args:        cobra.ExactArgs(1),
	Annotations: map[string]string{annotationRequiresService: "true"},
	RunE:        runDeleteBlocker,
}

func init() {
	deleteCmd.AddCommand(deleteBlockerCmd)
}

func runDeleteBlocker(cmd *cobra.Command, args []string) error {
	id, err := parseID(args[0], "blocker")
	if err != nil {
		return err
	}

	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	if _, err := proto.NewItemServiceClient(conn).DeleteBlocker(
		cmd.Context(),
		&proto.DeleteBlockerRequest{Id: id},
	); err != nil {
		return fmt.Errorf("failed to delete the blocker: %w", err)
	}

	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "deleted blocker %d\n", id); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}