package cmd

import (
	"fmt"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

type updateCommentOptions struct {
	Body string
}

var updateCommentOpts updateCommentOptions

// updateCommentCmd edits the body of an existing comment.
var updateCommentCmd = &cobra.Command{
	Use:         "comment [comment-id]",
	Short:       "Update a comment body",
	Example:     `  todo update comment 3 --body "still waiting on review"`,
	Args:        cobra.ExactArgs(1),
	Annotations: map[string]string{annotationRequiresService: "true"},
	RunE:        runUpdateComment,
}

func init() {
	updateCmd.AddCommand(updateCommentCmd)

	updateCommentCmd.Flags().StringVar(&updateCommentOpts.Body, "body", "", "New body of the comment")
	_ = updateCommentCmd.MarkFlagRequired("body")
}

func runUpdateComment(cmd *cobra.Command, args []string) error {
	id, err := parseID(args[0], "comment")
	if err != nil {
		return err
	}

	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	comment, err := proto.NewItemServiceClient(conn).UpdateComment(
		cmd.Context(),
		&proto.UpdateCommentRequest{Id: id, Body: updateCommentOpts.Body},
	)
	if err != nil {
		return fmt.Errorf("failed to update the comment: %w", err)
	}

	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "comment %d updated\n", comment.GetId()); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}