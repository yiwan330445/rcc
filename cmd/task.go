package cmd

import (
	"github.com/robocorp/rcc/common"
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
	if common.Product.IsLegacy() {
		rootCmd.AddCommand(taskCmd)

		taskCmd.PersistentFlags().BoolVarP(&common.ExternallyManaged, "externally-managed", "", false, "mark created Python environments as EXTERNALLY-MANAGED (PEP 668)")
	}
}
