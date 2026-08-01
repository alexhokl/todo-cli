package cmd

import (
	"fmt"

	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
)

const defaultLabelColour = "#FFFF00"

var createLabelColour string

// createLabelCmd creates a label explicitly.
var createLabelCmd = &cobra.Command{
	Use:   "label [name]",
	Short: "Create a label",
	Long: `Create a label.

Names are trimmed and stored in lower case, so "Work" and "work" are the same
label. Creating a label that already exists is reported as an error; use
` + "`todo update todo --add-label`" + ` to tag a todo, which creates missing
labels automatically.`,
	Example: `  todo create label urgent
  todo create label urgent --colour "#FF0000"`,
	Args:        cobra.ExactArgs(1),
	Annotations: map[string]string{annotationRequiresService: "true"},
	RunE:        runCreateLabel,
}

func init() {
	createCmd.AddCommand(createLabelCmd)
	createLabelCmd.Flags().StringVar(&createLabelColour, "colour", defaultLabelColour, "Colour code in #RRGGBB format")
}

func runCreateLabel(cmd *cobra.Command, args []string) error {
	conn, err := dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	label, err := proto.NewItemServiceClient(conn).CreateLabel(
		cmd.Context(),
		&proto.CreateLabelRequest{Name: args[0], Colour: createLabelColour},
	)
	if err != nil {
		return fmt.Errorf("failed to create the label: %w", err)
	}

	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s (id %d)\n", label.GetName(), label.GetId()); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}
