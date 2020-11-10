package cmd

import (
	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:     "env",
	Aliases: []string{"environment", "e"},
	Short:   "Group of commands related to `environment management`.",
	Long: `This "env" command set is for managing virtual environments
used in task context locally.`,
}

func init() {
	rootCmd.AddCommand(envCmd)
}
