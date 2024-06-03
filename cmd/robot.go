package cmd

import (
	"github.com/robocorp/rcc/common"
	"github.com/spf13/cobra"
)

var robotCmd = &cobra.Command{
	Use:     "robot",
	Aliases: []string{"r"},
	Short:   "Group of commands related to `robot`.",
	Long: `This set of commands relate to Robocorp Control Room related tasks. They are
executed either locally, or in connection to Robocorp Control Room and tooling.`,
}

func init() {
	if common.Product.IsLegacy() {
		rootCmd.AddCommand(robotCmd)
	}
}
