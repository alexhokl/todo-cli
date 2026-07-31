package cmd

import (
	"fmt"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

type renameLabelOptions struct {
	Name string
}

var renameLabelOpts renameLabelOptions

// renameLabelCmd renames an existing label. Note that this takes a label ID,
// unlike `todo update todo`, which takes a todo ID.
var renameLabelCmd = &cobra.Command{
	Use:         "label [id]",
	Short:       "Rename a label",
	Example:     `  todo update label 3 --name errands`,
	Args:        cobra.ExactArgs(1),
	Annotations: map[string]string{annotationRequiresService: "true"},
	RunE:        runRenameLabel,
}

func init() {
	updateCmd.AddCommand(renameLabelCmd)

	renameLabelCmd.Flags().StringVar(&renameLabelOpts.Name, "name", "", "New name of the label")
	_ = renameLabelCmd.MarkFlagRequired("name")
}

func runRenameLabel(cmd *cobra.Command, args []string) error {
	id, err := parseID(args[0], "label")
	if err != nil {
		return err
	}

	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	label, err := proto.NewItemServiceClient(conn).RenameLabel(
		cmd.Context(),
		&proto.RenameLabelRequest{Id: id, Name: renameLabelOpts.Name},
	)
	if err != nil {
		return fmt.Errorf("failed to rename the label: %w", err)
	}

	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s (id %d)\n", label.GetName(), label.GetId()); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}
