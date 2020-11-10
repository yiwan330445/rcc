package cmd

import (
	"github.com/spf13/cobra"
)

var feedbackCmd = &cobra.Command{
	Use:     "feedback",
	Aliases: []string{"f"},
	Short:   "Group of commands related to `rcc feedback`.",
	Long:    "Command group related to user feedback.",
	Hidden:  true,
}

func init() {
	rootCmd.AddCommand(feedbackCmd)
}
