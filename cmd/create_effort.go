package cmd

import (
	"fmt"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

// createEffortCmd creates an effort explicitly.
var createEffortCmd = &cobra.Command{
	Use:   "effort [name]",
	Short: "Create an effort",
	Long: `Create an effort.

Names are trimmed and stored in lower case, so "High" and "high" are the same
effort. Creating an effort that already exists is reported as an error; use
` + "`todo update effort`" + ` to attach an effort to a todo.`,
	Example:     `  todo create effort high`,
	Args:        cobra.ExactArgs(1),
	Annotations: map[string]string{annotationRequiresService: "true"},
	RunE:        runCreateEffort,
}

func init() {
	createCmd.AddCommand(createEffortCmd)
}

func runCreateEffort(cmd *cobra.Command, args []string) error {
	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	effort, err := proto.NewItemServiceClient(conn).CreateEffort(
		cmd.Context(),
		&proto.CreateEffortRequest{Name: args[0]},
	)
	if err != nil {
		return fmt.Errorf("failed to create the effort: %w", err)
	}

	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s (id %d)\n", effort.GetName(), effort.GetId()); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}