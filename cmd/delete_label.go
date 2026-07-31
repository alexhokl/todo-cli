package cmd

import (
	"fmt"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

// deleteLabelCmd deletes a label that is no longer in use.
var deleteLabelCmd = &cobra.Command{
	Use:   "label [id]",
	Short: "Delete a label",
	Long: `Delete a label.

A label that is still attached to a todo is not deleted; detach it first with
` + "`todo update todo --remove-label`" + `. This keeps deleting a label from
quietly changing the tagging of a todo.`,
	Example:     `  todo delete label 3`,
	Args:        cobra.ExactArgs(1),
	Annotations: map[string]string{annotationRequiresService: "true"},
	RunE:        runDeleteLabel,
}

func init() {
	deleteCmd.AddCommand(deleteLabelCmd)
}

func runDeleteLabel(cmd *cobra.Command, args []string) error {
	id, err := parseID(args[0], "label")
	if err != nil {
		return err
	}

	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	if _, err := proto.NewItemServiceClient(conn).DeleteLabel(
		cmd.Context(),
		&proto.DeleteLabelRequest{Id: id},
	); err != nil {
		return fmt.Errorf("failed to delete the label: %w", err)
	}

	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "deleted label %d\n", id); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}
