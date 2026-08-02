package cmd

import (
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

// listCommentsCmd lists every comment attached to an item.
var listCommentsCmd = &cobra.Command{
	Use:         "comments [item-id]",
	Aliases:     []string{"comment [item-id]"},
	Short:       "List comments on an item",
	Args:        cobra.ExactArgs(1),
	Annotations: map[string]string{annotationRequiresService: "true"},
	RunE:        runListComments,
}

func init() {
	listCmd.AddCommand(listCommentsCmd)
}

func runListComments(cmd *cobra.Command, args []string) error {
	id, err := parseID(args[0], "item")
	if err != nil {
		return err
	}

	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	response, err := proto.NewItemServiceClient(conn).ListComments(cmd.Context(), &proto.ListCommentsRequest{ItemId: id})
	if err != nil {
		return fmt.Errorf("failed to list the comments: %w", err)
	}

	out := cmd.OutOrStdout()
	if len(response.GetComments()) == 0 {
		if _, err := fmt.Fprintln(out, "  (none)"); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		return nil
	}

	writer := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(writer, "ID\tCREATED\tAUTHOR\tBODY"); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}
	for _, comment := range response.GetComments() {
		created := "-"
		if comment.GetCreatedAt() != nil {
			created = comment.GetCreatedAt().AsTime().Local().Format(time.DateTime)
		}
		if _, err := fmt.Fprintf(writer, "%d\t%s\t%s\t%s\n", comment.GetId(), created, comment.GetAuthor(), comment.GetBody()); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}