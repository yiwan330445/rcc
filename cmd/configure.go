package cmd

import (
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:     "configuration",
	Aliases: []string{"conf", "config", "configure"},
	Short:   "Group of commands related to `rcc configuration`.",
	Long:    "Group of commands to configure rcc with your settings.",
}

func init() {
	rootCmd.AddCommand(configureCmd)

	configureCmd.PersistentFlags().StringVarP(&accountName, "account", "a", "", "Account used for Robocorp Control Room task.")
}
