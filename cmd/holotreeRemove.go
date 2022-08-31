package cmd

import (
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/htfs"
	"github.com/robocorp/rcc/pretty"
	"github.com/spf13/cobra"
)

var (
	removeCheckRetries int
)

func holotreeRemove(catalogs []string) {
	common.Debug("Trying to remove following catalogs:")
	for _, catalog := range catalogs {
		common.Debug("- %s", catalog)
	}

	tree, err := htfs.New()
	pretty.Guard(err == nil, 2, "%s", err)

	err = tree.Remove(catalogs)
	pretty.Guard(err == nil, 3, "%s", err)
}

var holotreeRemoveCmd = &cobra.Command{
	Use:     "remove catalog+",
	Short:   "Remove existing holotree catalogs.",
	Long:    "Remove existing holotree catalogs. Partial identities are ok.",
	Aliases: []string{"rm"},
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if common.DebugFlag {
			defer common.Stopwatch("Holotree remove command lasted").Report()
		}
		if removeCheckRetries > 0 {
			checkLoop(removeCheckRetries)
		} else {
			pretty.Warning("Remember to run `rcc holotree check` after you have removed all desired catalogs!")
		}
		holotreeRemove(selectCatalogs(args))
		pretty.Ok()
	},
}

func init() {
	holotreeRemoveCmd.Flags().IntVarP(&removeCheckRetries, "check", "c", 0, "Additionally run holotree check with this many times.")
	holotreeCmd.AddCommand(holotreeRemoveCmd)
}
