package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "grit",
	Short: "A CLI tool for managing GitHub issues",
	Long:  "grit allows you to create, close, comment on, and manage GitHub issues from the command line.",
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("grit %s\n", version)
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
