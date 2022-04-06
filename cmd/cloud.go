package cmd

import (
	"github.com/spf13/cobra"
)

var cloudCmd = &cobra.Command{
	Use:     "cloud",
	Aliases: []string{"robocorp", "c"},
	Short:   "Group of commands related to `Robocorp Control Room`.",
	Long:    `This group of commands apply to communication with Robocorp Control Room.`,
}

func init() {
	rootCmd.AddCommand(cloudCmd)

	cloudCmd.PersistentFlags().StringVarP(&accountName, "account", "a", "", "Account used for Robocorp Control Room operations.")
}
