package cmd

import (
	"fmt"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

type renameEffortOptions struct {
	Name string
}

var renameEffortOpts renameEffortOptions

// renameEffortCmd renames an existing effort. Note that this takes an effort
// ID, unlike `todo update todo`, which takes a todo ID.
var renameEffortCmd = &cobra.Command{
	Use:         "effort [id]",
	Short:       "Rename an effort",
	Example:     `  todo update effort 3 --name medium`,
	Args:        cobra.ExactArgs(1),
	Annotations: map[string]string{annotationRequiresService: "true"},
	RunE:        runRenameEffort,
}

func init() {
	updateCmd.AddCommand(renameEffortCmd)

	renameEffortCmd.Flags().StringVar(&renameEffortOpts.Name, "name", "", "New name of the effort")
	_ = renameEffortCmd.MarkFlagRequired("name")
}

func runRenameEffort(cmd *cobra.Command, args []string) error {
	id, err := parseID(args[0], "effort")
	if err != nil {
		return err
	}

	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	effort, err := proto.NewItemServiceClient(conn).RenameEffort(
		cmd.Context(),
		&proto.RenameEffortRequest{Id: id, Name: renameEffortOpts.Name},
	)
	if err != nil {
		return fmt.Errorf("failed to rename the effort: %w", err)
	}

	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s (id %d)\n", effort.GetName(), effort.GetId()); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}