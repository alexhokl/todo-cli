package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a subject",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println("Please specify the subject to be created")
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}
