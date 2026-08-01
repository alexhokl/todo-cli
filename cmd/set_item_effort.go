package cmd

import (
	"fmt"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

type setItemEffortOptions struct {
	Effort     string
	ClearEffort bool
}

var setItemEffortOpts setItemEffortOptions

// setItemEffortCmd attaches an effort to an item by name, or clears it.
var setItemEffortCmd = &cobra.Command{
	Use:   "todo [id]",
	Short: "Set or clear the effort of an item",
	Long: `Set the effort of an item by name. The effort must already exist; use
` + "`todo create effort`" + ` first. Pass --clear-effort to detach the effort
instead.`,
	Example: `  todo update todo 7 --effort high
  todo update todo 7 --clear-effort`,
	Args:        cobra.ExactArgs(1),
	Annotations: map[string]string{annotationRequiresService: "true"},
	RunE:        runSetItemEffort,
}

func init() {
	updateCmd.AddCommand(setItemEffortCmd)

	flags := setItemEffortCmd.Flags()
	flags.StringVar(&setItemEffortOpts.Effort, "effort", "", "Name of the effort to attach")
	flags.BoolVar(&setItemEffortOpts.ClearEffort, "clear-effort", false, "Detach the effort from the item")

	setItemEffortCmd.MarkFlagsMutuallyExclusive("effort", "clear-effort")
}

func runSetItemEffort(cmd *cobra.Command, args []string) error {
	id, err := parseID(args[0], "item")
	if err != nil {
		return err
	}

	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	effort := setItemEffortOpts.Effort
	if setItemEffortOpts.ClearEffort {
		effort = ""
	}

	item, err := proto.NewItemServiceClient(conn).SetItemEffort(
		cmd.Context(),
		&proto.SetItemEffortRequest{Id: id, Effort: effort},
	)
	if err != nil {
		return fmt.Errorf("failed to set the effort: %w", err)
	}

	return writeItemLine(cmd.OutOrStdout(), item)
}