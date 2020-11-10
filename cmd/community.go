package cmd

import (
	"github.com/spf13/cobra"
)

var communityCmd = &cobra.Command{
	Use:     "community",
	Aliases: []string{"co"},
	Short:   "Group of commands related to `Robocorp Community`.",
	Long:    `This group of commands apply to community provided robots and services.`,
}

func init() {
	rootCmd.AddCommand(communityCmd)
}
