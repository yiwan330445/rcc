package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/htfs"
	"github.com/robocorp/rcc/pretty"
	"github.com/spf13/cobra"
)

const mega = 1024 * 1024

func listCatalogDetails() {
	tabbed := tabwriter.NewWriter(os.Stderr, 2, 4, 2, ' ', 0)
	tabbed.Write([]byte("Blueprint\tPlatform\tDirs  \tFiles  \tSize   \tidentity.yaml (gzipped blob inside ROBOCORP_HOME)\tHolotree path\n"))
	tabbed.Write([]byte("---------\t--------\t------\t-------\t-------\t-------------------------------------------------\t-------------\n"))
	_, roots := htfs.LoadCatalogs()
	for _, catalog := range roots {
		stats, err := catalog.Stats()
		pretty.Guard(err == nil, 1, "Could not get stats for %s, reason: %s", catalog.Blueprint, err)
		megas := stats.Bytes / mega
		data := fmt.Sprintf("%s\t%s\t% 6d\t% 7d\t% 6dM\t%s\t%s\n", catalog.Blueprint, catalog.Platform, stats.Directories, stats.Files, megas, stats.Identity, catalog.HolotreeBase())
		tabbed.Write([]byte(data))
	}
	tabbed.Flush()
}

var holotreeCatalogsCmd = &cobra.Command{
	Use:   "catalogs",
	Short: "List native and imported holotree catalogs.",
	Long:  "List native and imported holotree catalogs.",
	Run: func(cmd *cobra.Command, args []string) {
		if common.DebugFlag {
			defer common.Stopwatch("Holotree catalogs command lasted").Report()
		}
		listCatalogDetails()
		pretty.Ok()
	},
}

func init() {
	holotreeCmd.AddCommand(holotreeCatalogsCmd)
}
