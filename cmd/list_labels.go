package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

// listLabelsCmd lists every known label.
var listLabelsCmd = &cobra.Command{
	Use:         "labels",
	Short:       "List labels",
	Args:        cobra.NoArgs,
	Annotations: map[string]string{annotationRequiresService: "true"},
	RunE:        runListLabels,
}

func init() {
	listCmd.AddCommand(listLabelsCmd)
}

func runListLabels(cmd *cobra.Command, _ []string) error {
	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	response, err := proto.NewItemServiceClient(conn).ListLabels(cmd.Context(), &proto.ListLabelsRequest{})
	if err != nil {
		return fmt.Errorf("failed to list the labels: %w", err)
	}

	out := cmd.OutOrStdout()
	if len(response.GetLabels()) == 0 {
		if _, err := fmt.Fprintln(out, "  (none)"); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		return nil
	}

	writer := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(writer, "ID\tNAME"); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}
	for _, label := range response.GetLabels() {
		if _, err := fmt.Fprintf(writer, "%d\t%s\n", label.GetId(), label.GetName()); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}
