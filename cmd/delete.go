package cmd

import (
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete one managed virtual environment.",
	Long: `Delete the given virtual environment from existence.
After deletion, it will not be available anymore.`,
	Run: func(cmd *cobra.Command, args []string) {
		for _, label := range args {
			common.Log("Removing %v", label)
			conda.RemoveEnvironment(label)
		}
	},
}

func init() {
	envCmd.AddCommand(deleteCmd)
}
