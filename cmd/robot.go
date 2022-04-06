package cmd

import (
	"github.com/spf13/cobra"
)

var robotCmd = &cobra.Command{
	Use:     "robot",
	Aliases: []string{"r"},
	Short:   "Group of commands related to `robot`.",
	Long: `This set of commands relate to Robocorp Control Room related tasks. They are
executed either locally, or in connection to Robocorp Control Room and Robocorp App.`,
}

func init() {
	rootCmd.AddCommand(robotCmd)
}
