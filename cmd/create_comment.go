package cmd

import (
	"fmt"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

type createCommentOptions struct {
	Body string
}

var createCommentOpts createCommentOptions

// createCommentCmd attaches a new comment to an item.
var createCommentCmd = &cobra.Command{
	Use:         "comment [item-id]",
	Short:       "Create a comment on an item",
	Example:     `  todo create comment 7 --body "drafted a reply, waiting on review"`,
	Args:        cobra.ExactArgs(1),
	Annotations: map[string]string{annotationRequiresService: "true"},
	RunE:        runCreateComment,
}

func init() {
	createCmd.AddCommand(createCommentCmd)

	createCommentCmd.Flags().StringVar(&createCommentOpts.Body, "body", "", "Body of the comment")
	_ = createCommentCmd.MarkFlagRequired("body")
}

func runCreateComment(cmd *cobra.Command, args []string) error {
	itemID, err := parseID(args[0], "item")
	if err != nil {
		return err
	}

	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	comment, err := proto.NewItemServiceClient(conn).CreateComment(
		cmd.Context(),
		&proto.CreateCommentRequest{ItemId: itemID, Body: createCommentOpts.Body},
	)
	if err != nil {
		return fmt.Errorf("failed to create the comment: %w", err)
	}

	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "comment %d on item %d\n", comment.GetId(), itemID); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}