package cmd

import (
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:     "configure",
	Aliases: []string{"conf", "config"},
	Short:   "Group of commands related to `rcc configuration`.",
	Long:    "Group of commands to configure rcc with your settings.",
}

func init() {
	rootCmd.AddCommand(configureCmd)

	configureCmd.PersistentFlags().StringVarP(&accountName, "account", "a", "", "Account used for Robocorp Cloud task.")
}
