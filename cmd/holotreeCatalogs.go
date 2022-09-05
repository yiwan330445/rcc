package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/htfs"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/pretty"
	"github.com/spf13/cobra"
)

var (
	showIdentityYaml bool
)

const mega = 1024 * 1024

func megas(bytes uint64) uint64 {
	return bytes / mega
}

func catalogUsedStats() map[string]int {
	result := make(map[string]int)
	handle, err := os.Open(common.HololibUsageLocation())
	if err != nil {
		return result
	}
	defer handle.Close()
	entries, err := handle.Readdir(-1)
	if err != nil {
		return result
	}
	for _, entry := range entries {
		name := filepath.Base(entry.Name())
		tail := filepath.Ext(name)
		size := len(name) - len(tail)
		base := name[:size]
		days := common.DayCountSince(entry.ModTime())
		previous, ok := result[base]
		if !ok || days < previous {
			result[base] = days
		}
	}
	return result
}

func identityContent(catalog *htfs.Root) string {
	blob, err := catalog.Show("identity.yaml")
	if err != nil {
		return err.Error()
	}
	return string(blob)
}

func identityContentLines(catalog *htfs.Root) []string {
	content := identityContent(catalog)
	result := strings.SplitAfter(content, "\n")
	for at, value := range result {
		result[at] = strings.Replace(strings.TrimRight(value, "\r\n\t "), "\t", "  ", -1)
	}
	return result
}

func jsonCatalogDetails(roots []*htfs.Root) {
	used := catalogUsedStats()
	holder := make(map[string]map[string]interface{})
	for _, catalog := range roots {
		lastUse, ok := used[catalog.Blueprint]
		if !ok {
			catalog.Touch()
			lastUse = -1
		}
		stats, err := catalog.Stats()
		pretty.Guard(err == nil, 1, "Could not get stats for %s, reason: %s", catalog.Blueprint, err)
		data := make(map[string]interface{})
		data["blueprint"] = catalog.Blueprint
		data["holotree"] = catalog.HolotreeBase()
		identity := filepath.Join(common.HololibLibraryLocation(), stats.Identity)
		data["identity.yaml"] = identity
		if showIdentityYaml {
			data["identity-content"] = identityContent(catalog)
		}
		data["platform"] = catalog.Platform
		data["directories"] = stats.Directories
		data["files"] = stats.Files
		data["bytes"] = stats.Bytes
		holder[catalog.Blueprint] = data
		age, _ := pathlib.DaysSinceModified(catalog.Source())
		data["age_in_days"] = age
		data["days_since_last_use"] = lastUse
	}
	nice, err := json.MarshalIndent(holder, "", "  ")
	pretty.Guard(err == nil, 2, "%s", err)
	common.Stdout("%s\n", nice)
}

func listCatalogDetails(roots []*htfs.Root) {
	used := catalogUsedStats()
	tabbed := tabwriter.NewWriter(os.Stderr, 2, 4, 2, ' ', 0)
	tabbed.Write([]byte("Blueprint\tPlatform\tDirs  \tFiles  \tSize   \tidentity.yaml (gzipped blob inside hololib)\tHolotree path\tAge (days)\tIdle (days)\n"))
	tabbed.Write([]byte("---------\t--------\t------\t-------\t-------\t-------------------------------------------\t-------------\t----------\t-----------\n"))
	for _, catalog := range roots {
		lastUse, ok := used[catalog.Blueprint]
		if !ok {
			catalog.Touch()
			lastUse = -1
		}
		stats, err := catalog.Stats()
		pretty.Guard(err == nil, 1, "Could not get stats for %s, reason: %s", catalog.Blueprint, err)
		days, _ := pathlib.DaysSinceModified(catalog.Source())
		data := fmt.Sprintf("%s\t%s\t% 6d\t% 7d\t% 6dM\t%s\t%s\t%10d\t%11d\n", catalog.Blueprint, catalog.Platform, stats.Directories, stats.Files, megas(stats.Bytes), stats.Identity, catalog.HolotreeBase(), days, lastUse)
		tabbed.Write([]byte(data))
		if showIdentityYaml {
			for _, line := range identityContentLines(catalog) {
				tabbed.Write([]byte(fmt.Sprintf("\t\t\t\t\t%s\n", line)))
			}
		}
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
	holotreeCatalogsCmd.Flags().BoolVarP(&showIdentityYaml, "identity", "i", false, "Show identity.yaml in catalog context.")
}
