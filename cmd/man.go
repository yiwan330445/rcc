package cmd

import (
	"github.com/spf13/cobra"
)

var manCmd = &cobra.Command{
	Use:     "man",
	Aliases: []string{"manuals", "docs", "doc", "guides", "guide", "m"},
	Short:   "Group of commands related to `rcc documentation`.",
	Long:    "Build in documentation and manuals.",
}

func init() {
	rootCmd.AddCommand(manCmd)
}
