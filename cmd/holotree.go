package cmd

import (
	"github.com/spf13/cobra"
)

var holotreeCmd = &cobra.Command{
	Use:     "holotree",
	Aliases: []string{"ht"},
	Short:   "Group of holotree commands.",
	Long:    "Group of holotree commands.",
	Hidden:  true,
}

func init() {
	rootCmd.AddCommand(holotreeCmd)
}
