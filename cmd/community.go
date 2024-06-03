package cmd

import (
	"github.com/robocorp/rcc/common"
	"github.com/spf13/cobra"
)

var communityCmd = &cobra.Command{
	Use:     "community",
	Aliases: []string{"co"},
	Short:   "Group of commands related to `Robocorp Community`.",
	Long:    `This group of commands apply to community provided robots and services.`,
}

func init() {
	if common.Product.IsLegacy() {
		rootCmd.AddCommand(communityCmd)
	}
}
