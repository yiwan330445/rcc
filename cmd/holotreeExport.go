package cmd

import (
	"strings"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/htfs"
	"github.com/robocorp/rcc/pretty"
	"github.com/spf13/cobra"
)

var (
	holozip string
)

func holotreeExport(catalogs []string, archive string) {
	common.Debug("Exporting catalogs:")
	for _, catalog := range catalogs {
		common.Debug("- %s", catalog)
	}

	tree, err := htfs.New()
	pretty.Guard(err == nil, 2, "%s", err)

	err = tree.Export(catalogs, archive)
	pretty.Guard(err == nil, 3, "%s", err)
}

func listCatalogs() {
	common.Log("Selectable catalogs (you can use substrings):")
	for _, catalog := range htfs.Catalogs() {
		common.Log("- %s", catalog)
	}
}

func selectCatalogs(filters []string) []string {
	result := make([]string, 0, len(filters))
	for _, catalog := range htfs.Catalogs() {
		for _, filter := range filters {
			if strings.Contains(catalog, filter) {
				result = append(result, catalog)
				break
			}
		}
	}
	return result
}

var holotreeExportCmd = &cobra.Command{
	Use:   "export catalog+",
	Short: "Export existing holotree catalog and library parts.",
	Long:  "Export existing holotree catalog and library parts.",
	Run: func(cmd *cobra.Command, args []string) {
		if common.DebugFlag {
			defer common.Stopwatch("Holotree export command lasted").Report()
		}
		if len(args) == 0 {
			listCatalogs()
		} else {
			holotreeExport(selectCatalogs(args), holozip)
		}
		pretty.Ok()
	},
}

func init() {
	holotreeCmd.AddCommand(holotreeExportCmd)
	holotreeExportCmd.Flags().StringVarP(&holozip, "zipfile", "z", "hololib.zip", "Name of zipfile to export.")
}
