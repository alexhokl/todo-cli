package cmd

import (
	"fmt"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

// deleteCommentCmd deletes a comment.
var deleteCommentCmd = &cobra.Command{
	Use:         "comment [comment-id]",
	Short:       "Delete a comment",
	Example:     `  todo delete comment 3`,
	Args:        cobra.ExactArgs(1),
	Annotations: map[string]string{annotationRequiresService: "true"},
	RunE:        runDeleteComment,
}

func init() {
	deleteCmd.AddCommand(deleteCommentCmd)
}

func runDeleteComment(cmd *cobra.Command, args []string) error {
	id, err := parseID(args[0], "comment")
	if err != nil {
		return err
	}

	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	if _, err := proto.NewItemServiceClient(conn).DeleteComment(
		cmd.Context(),
		&proto.DeleteCommentRequest{Id: id},
	); err != nil {
		return fmt.Errorf("failed to delete the comment: %w", err)
	}

	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "deleted comment %d\n", id); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}