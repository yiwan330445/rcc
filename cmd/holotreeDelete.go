package cmd

import (
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/htfs"
	"github.com/robocorp/rcc/pretty"

	"github.com/spf13/cobra"
)

var holotreeDeleteCmd = &cobra.Command{
	Use:   "delete <partial identity>+",
	Short: "Delete holotree controller space.",
	Long:  "Delete holotree controller space.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		for _, prefix := range args {
			for _, label := range htfs.FindEnvironment(prefix) {
				common.Log("Removing %v", label)
				if dryFlag {
					continue
				}
				err := htfs.RemoveHolotreeSpace(label)
				pretty.Guard(err == nil, 1, "Error: %v", err)
			}
		}
	},
}

func init() {
	holotreeCmd.AddCommand(holotreeDeleteCmd)
	holotreeDeleteCmd.Flags().BoolVarP(&dryFlag, "dryrun", "d", false, "Don't delete environments, just show what would happen.")
}
