package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/htfs"
	"github.com/robocorp/rcc/pretty"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	holozip     string
	exportRobot string
	specFile    string
)

type (
	ExportSpec struct {
		Workspace string   `yaml:"workspace"`
		Known     []string `yaml:"knows"`
		Needs     []string `yaml:"wants"`
	}
)

func specFrom(content []byte) (*ExportSpec, error) {
	result := &ExportSpec{}
	err := yaml.Unmarshal(content, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func loadExportSpec(filename string) (*ExportSpec, error) {
	raw, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	spec, err := specFrom(raw)
	if err != nil {
		return nil, err
	}
	return spec, nil
}

func exportBySpecification(filename string) {
	spec, err := loadExportSpec(filename)
	pretty.Guard(err == nil, 4, "Loading specification %q failed, reason: %v", filename, err)
	known := selectExactCatalogs(spec.Known)
	needs := selectExactCatalogs(spec.Needs)
	pretty.Guard(len(spec.Needs) == len(needs), 5, "Only %d out of %d needed catalogs available. Quitting!", len(needs), len(spec.Needs))
	unifiedSpec := &ExportSpec{
		Workspace: spec.Workspace,
		Known:     known,
		Needs:     needs,
	}
	content, err := yaml.Marshal(unifiedSpec)
	pretty.Guard(err == nil, 6, "Marshaling unified specification failed, reason: %v", err)
	fingerprint := common.Siphash(9007199254740993, 2147483647, content)
	common.Debug("Final delta specification %0x16x is:\n%s", fingerprint, string(content))
	deltafile := fmt.Sprintf("%016x.hld", fingerprint)
	holotreeExport(needs, known, deltafile)
	common.Stdout("%s\n", deltafile)
}

func holotreeExport(catalogs, known []string, archive string) {
	common.Debug("Ignoring content from catalogs:")
	for _, catalog := range known {
		common.Debug("- %s", catalog)
	}

	common.Debug("Exporting catalogs:")
	for _, catalog := range catalogs {
		common.Debug("- %s", catalog)
	}

	tree, err := htfs.New()
	pretty.Guard(err == nil, 2, "%s", err)

	err = tree.Export(catalogs, known, archive)
	pretty.Guard(err == nil, 3, "%s", err)
}

func listCatalogs(jsonForm bool) {
	if jsonForm {
		nice, err := json.MarshalIndent(htfs.Catalogs(), "", "  ")
		pretty.Guard(err == nil, 2, "%s", err)
		common.Stdout("%s\n", nice)
	} else {
		common.Log("Selectable catalogs (you can use substrings):")
		for _, catalog := range htfs.Catalogs() {
			common.Log("- %s", catalog)
		}
	}
}

func selectExactCatalogs(filters []string) []string {
	result := make([]string, 0, len(filters))
	for _, catalog := range htfs.Catalogs() {
		for _, filter := range filters {
			if catalog == filter {
				result = append(result, catalog)
				break
			}
		}
	}
	sort.Strings(result)
	return result
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
	sort.Strings(result)
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
		if len(specFile) > 0 {
			exportBySpecification(specFile)
			return
		}
		if len(exportRobot) > 0 {
			_, holotreeBlueprint, err := htfs.ComposeFinalBlueprint(nil, exportRobot)
			pretty.Guard(err == nil, 1, "Blueprint calculation failed: %v", err)
			hash := htfs.BlueprintHash(holotreeBlueprint)
			args = append(args, htfs.CatalogName(hash))
		}
		if len(args) == 0 {
			listCatalogs(jsonFlag)
		} else {
			holotreeExport(selectCatalogs(args), nil, holozip)
		}
		pretty.Ok()
	},
}

func init() {
	holotreeCmd.AddCommand(holotreeExportCmd)
	holotreeExportCmd.Flags().StringVarP(&specFile, "specification", "s", "", "Filename to use as export speficifaction in YAML format.")
	holotreeExportCmd.Flags().StringVarP(&holozip, "zipfile", "z", "hololib.zip", "Name of zipfile to export.")
	holotreeExportCmd.Flags().BoolVarP(&jsonFlag, "json", "j", false, "Output in JSON format")
	holotreeExportCmd.Flags().StringVarP(&exportRobot, "robot", "r", "", "Full path to 'robot.yaml' configuration file to export as catalog. <optional>")
}
