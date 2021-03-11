package cmd

import (
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/pretty"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete one managed virtual environment.",
	Long: `Delete the given virtual environment from existence.
After deletion, it will not be available anymore.`,
	Run: func(cmd *cobra.Command, args []string) {
		for _, prefix := range args {
			for _, label := range conda.FindEnvironment(prefix) {
				common.Log("Removing %v", label)
				err := conda.RemoveEnvironment(label)
				pretty.Guard(err == nil, 1, "Error: %v", err)
			}
		}
	},
}

func init() {
	envCmd.AddCommand(deleteCmd)
}
