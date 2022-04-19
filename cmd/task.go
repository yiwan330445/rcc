package cmd

import (
	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:     "task",
	Aliases: []string{"t"},
	Short:   "Group of commands related to `task`.",
	Long: `This set of commands relate to Robocorp Control Room related tasks. They are
executed either locally, or in connection to Robocorp Control Room and tooling.`,
}

func init() {
	rootCmd.AddCommand(taskCmd)
}
