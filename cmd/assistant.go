package cmd

import (
	"github.com/robocorp/rcc/common"
	"github.com/spf13/cobra"
)

var assistantCmd = &cobra.Command{
	Use:     "assistant",
	Aliases: []string{"assist", "a"},
	Short:   "Group of commands related to `robot assistant`.",
	Long: `This set of commands relate to Robocorp Robot Assistant related tasks.
They are either local, or in relation to Robocorp Control Room and tooling.`,
}

func init() {
	if common.Product.IsLegacy() {
		rootCmd.AddCommand(assistantCmd)

		assistantCmd.PersistentFlags().StringVarP(&accountName, "account", "", "", "Account used for Robocorp Control Room operations.")
	}
}
