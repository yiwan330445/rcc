package cmd

import (
	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:     "task",
	Aliases: []string{"t"},
	Short:   "Group of commands related to `task`.",
	Long: `This set of commands relate to Robocorp Cloud related tasks. They are
executed either locally, or in connection to Robocorp Cloud and Robocorp App.`,
}

func init() {
	rootCmd.AddCommand(taskCmd)
}
