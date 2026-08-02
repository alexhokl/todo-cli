package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

// listEffortsCmd lists every known effort.
var listEffortsCmd = &cobra.Command{
	Use:         "efforts",
	Aliases:     []string{"effort"},
	Short:       "List efforts",
	Args:        cobra.NoArgs,
	Annotations: map[string]string{annotationRequiresService: "true"},
	RunE:        runListEfforts,
}

func init() {
	listCmd.AddCommand(listEffortsCmd)
}

func runListEfforts(cmd *cobra.Command, _ []string) error {
	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	response, err := proto.NewItemServiceClient(conn).ListEfforts(cmd.Context(), &proto.ListEffortsRequest{})
	if err != nil {
		return fmt.Errorf("failed to list the efforts: %w", err)
	}

	out := cmd.OutOrStdout()
	if len(response.GetEfforts()) == 0 {
		if _, err := fmt.Fprintln(out, "  (none)"); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		return nil
	}

	writer := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(writer, "ID\tNAME"); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}
	for _, effort := range response.GetEfforts() {
		if _, err := fmt.Fprintf(writer, "%d\t%s\n", effort.GetId(), effort.GetName()); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}