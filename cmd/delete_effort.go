package cmd

import (
	"fmt"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

// deleteEffortCmd deletes an effort that is no longer in use.
var deleteEffortCmd = &cobra.Command{
	Use:   "effort [id]",
	Short: "Delete an effort",
	Long: `Delete an effort.

An effort that is still attached to a todo is not deleted; detach it first with
` + "`todo update todo --clear-effort`" + `. This keeps deleting an effort from
quietly changing the effort of a todo.`,
	Example:     `  todo delete effort 3`,
	Args:        cobra.ExactArgs(1),
	Annotations: map[string]string{annotationRequiresService: "true"},
	RunE:        runDeleteEffort,
}

func init() {
	deleteCmd.AddCommand(deleteEffortCmd)
}

func runDeleteEffort(cmd *cobra.Command, args []string) error {
	id, err := parseID(args[0], "effort")
	if err != nil {
		return err
	}

	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	if _, err := proto.NewItemServiceClient(conn).DeleteEffort(
		cmd.Context(),
		&proto.DeleteEffortRequest{Id: id},
	); err != nil {
		return fmt.Errorf("failed to delete the effort: %w", err)
	}

	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "deleted effort %d\n", id); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}