package cmd

import (
	"github.com/spf13/cobra"
)

var interactiveCmd = &cobra.Command{
	Use:     "interactive",
	Aliases: []string{"i"},
	Short:   "Group of interactive commands. For human users. Do not use in automation.",
	Long: `This group of commands are interactive, asking questions from user when needed.
Do not try to use these in automation, they will fail there.`,
}

func init() {
	rootCmd.AddCommand(interactiveCmd)
}
