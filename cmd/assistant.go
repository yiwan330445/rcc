package cmd

import (
	"github.com/spf13/cobra"
)

var assistantCmd = &cobra.Command{
	Use:     "assistant",
	Aliases: []string{"assist", "a"},
	Short:   "Group of commands related to `robot assistant`.",
	Long: `This set of commands relate to Robocorp Robot Assistant related tasks.
They are either local, or in relation to Robocorp Cloud and Robocorp App.`,
}

func init() {
	rootCmd.AddCommand(assistantCmd)

	assistantCmd.PersistentFlags().StringVarP(&accountName, "account", "", "", "Account used for Robocorp Cloud operations.")
}
