package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a subject",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println("Please specify the subject to be retrieved")
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
