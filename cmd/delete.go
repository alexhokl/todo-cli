package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a subject",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println("Please specify the subject to be deleted")
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
