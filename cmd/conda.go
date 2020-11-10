package cmd

import (
	"github.com/spf13/cobra"
)

var condaCmd = &cobra.Command{
	Use:   "conda",
	Short: "Group of commands related to `conda installation`.",
	Long:  `Conda specific funtionality captured in this set of subcommands.`,
}

func init() {
	rootCmd.AddCommand(condaCmd)
}
