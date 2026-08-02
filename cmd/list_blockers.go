package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

// listBlockersCmd lists every blocker attached to an item.
var listBlockersCmd = &cobra.Command{
	Use:         "blockers [item-id]",
	Aliases:     []string{"blocker [item-id]"},
	Short:       "List blockers on an item",
	Args:        cobra.ExactArgs(1),
	Annotations: map[string]string{annotationRequiresService: "true"},
	RunE:        runListBlockers,
}

func init() {
	listCmd.AddCommand(listBlockersCmd)
}

func runListBlockers(cmd *cobra.Command, args []string) error {
	id, err := parseID(args[0], "item")
	if err != nil {
		return err
	}

	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	response, err := proto.NewItemServiceClient(conn).ListBlockers(cmd.Context(), &proto.ListBlockersRequest{ItemId: id})
	if err != nil {
		return fmt.Errorf("failed to list the blockers: %w", err)
	}

	out := cmd.OutOrStdout()
	if len(response.GetBlockers()) == 0 {
		if _, err := fmt.Fprintln(out, "  (none)"); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		return nil
	}

	writer := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(writer, "ID\tDESCRIPTION"); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}
	for _, blocker := range response.GetBlockers() {
		if _, err := fmt.Fprintf(writer, "%d\t%s\n", blocker.GetId(), blocker.GetDescription()); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}