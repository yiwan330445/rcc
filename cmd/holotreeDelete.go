package cmd

import (
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/htfs"
	"github.com/robocorp/rcc/pretty"

	"github.com/spf13/cobra"
)

func deleteByPartialIdentity(partials []string) {
	for _, prefix := range partials {
		for _, label := range htfs.FindEnvironment(prefix) {
			common.Log("Removing %v", label)
			if dryFlag {
				continue
			}
			err := htfs.RemoveHolotreeSpace(label)
			pretty.Guard(err == nil, 1, "Error: %v", err)
		}
	}
}

var holotreeDeleteCmd = &cobra.Command{
	Use:     "delete <partial identity>*",
	Short:   "Delete holotree controller space.",
	Long:    "Delete holotree controller space.",
	Aliases: []string{"del"},
	Run: func(cmd *cobra.Command, args []string) {
		partials := make([]string, 0, len(args)+1)
		if len(args) > 0 {
			partials = append(partials, args...)
		}
		if len(common.HolotreeSpace) > 0 {
			partials = append(partials, htfs.ControllerSpaceName([]byte(common.ControllerIdentity()), []byte(common.HolotreeSpace)))
		}
		pretty.Guard(len(partials) > 0, 1, "Must provide either --space flag, or partial environment identity!")
		deleteByPartialIdentity(partials)
	},
}

func init() {
	holotreeCmd.AddCommand(holotreeDeleteCmd)
	holotreeDeleteCmd.Flags().BoolVarP(&dryFlag, "dryrun", "d", false, "Don't delete environments, just show what would happen.")
	holotreeDeleteCmd.Flags().StringVarP(&common.HolotreeSpace, "space", "s", "user", "Client specific name to identify environment to delete.")
}
