package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/htfs"
	"github.com/robocorp/rcc/pretty"
	"github.com/spf13/cobra"
)

const mega = 1024 * 1024

func megas(bytes uint64) uint64 {
	return bytes / mega
}

func jsonCatalogDetails(roots []*htfs.Root) {
	holder := make(map[string]map[string]interface{})
	for _, catalog := range roots {
		stats, err := catalog.Stats()
		pretty.Guard(err == nil, 1, "Could not get stats for %s, reason: %s", catalog.Blueprint, err)
		data := make(map[string]interface{})
		data["blueprint"] = catalog.Blueprint
		data["holotree"] = catalog.HolotreeBase()
		data["identity.yaml"] = filepath.Join(common.RobocorpHome(), stats.Identity)
		data["platform"] = catalog.Platform
		data["directories"] = stats.Directories
		data["files"] = stats.Files
		data["bytes"] = stats.Bytes
		holder[catalog.Blueprint] = data
	}
	nice, err := json.MarshalIndent(holder, "", "  ")
	pretty.Guard(err == nil, 2, "%s", err)
	common.Stdout("%s\n", nice)
}

func listCatalogDetails(roots []*htfs.Root) {
	tabbed := tabwriter.NewWriter(os.Stderr, 2, 4, 2, ' ', 0)
	tabbed.Write([]byte("Blueprint\tPlatform\tDirs  \tFiles  \tSize   \tidentity.yaml (gzipped blob inside ROBOCORP_HOME)\tHolotree path\n"))
	tabbed.Write([]byte("---------\t--------\t------\t-------\t-------\t-------------------------------------------------\t-------------\n"))
	for _, catalog := range roots {
		stats, err := catalog.Stats()
		pretty.Guard(err == nil, 1, "Could not get stats for %s, reason: %s", catalog.Blueprint, err)
		data := fmt.Sprintf("%s\t%s\t% 6d\t% 7d\t% 6dM\t%s\t%s\n", catalog.Blueprint, catalog.Platform, stats.Directories, stats.Files, megas(stats.Bytes), stats.Identity, catalog.HolotreeBase())
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
		_, roots := htfs.LoadCatalogs()
		if jsonFlag {
			jsonCatalogDetails(roots)
		} else {
			listCatalogDetails(roots)
		}
		pretty.Ok()
	},
}

func init() {
	holotreeCmd.AddCommand(holotreeCatalogsCmd)
	holotreeCatalogsCmd.Flags().BoolVarP(&jsonFlag, "json", "j", false, "Output in JSON format")
}
