package cmd

import (
	"github.com/robocorp/rcc/anywork"
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/htfs"
	"github.com/robocorp/rcc/pretty"
	"github.com/spf13/cobra"
)

var (
	holotreeBlueprint []byte
	holotreeSpace     string
	holotreeForce     bool
	holotreeJson      bool
)

var holotreeVariablesCmd = &cobra.Command{
	Use:     "variables conda.yaml+",
	Aliases: []string{"vars"},
	Short:   "Do holotree operations.",
	Long:    "Do holotree operations.",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if common.DebugFlag {
			defer common.Stopwatch("Holotree command lasted").Report()
		}

		var left, right *conda.Environment
		var err error

		for _, filename := range args {
			left = right
			right, err = conda.ReadCondaYaml(filename)
			pretty.Guard(err == nil, 1, "Failure: %v", err)
			if left == nil {
				continue
			}
			right, err = left.Merge(right)
			pretty.Guard(err == nil, 1, "Failure: %v", err)
		}
		pretty.Guard(right != nil, 1, "Missing environment specification(s).")
		content, err := right.AsYaml()
		pretty.Guard(err == nil, 1, "YAML error: %v", err)
		holotreeBlueprint = []byte(content)

		ok := conda.MustMicromamba()
		pretty.Guard(ok, 1, "Could not get micromamba installed.")

		anywork.Scale(200)

		tree, err := htfs.RecordEnvironment(holotreeBlueprint, holotreeForce)
		pretty.Guard(err == nil, 2, "%w", err)

		path, err := tree.Restore(holotreeBlueprint, []byte(common.ControllerIdentity()), []byte(holotreeSpace))
		pretty.Guard(err == nil, 10, "Failed to restore blueprint %q, reason: %v", string(holotreeBlueprint), err)

		env := conda.EnvironmentExtensionFor(path)
		if holotreeJson {
			asJson(env)
		} else {
			asExportedText(env)
		}
	},
}

func init() {
	holotreeCmd.AddCommand(holotreeVariablesCmd)
	holotreeVariablesCmd.Flags().StringVar(&holotreeSpace, "space", "", "Client specific name to identify this environment.")
	holotreeVariablesCmd.MarkFlagRequired("space")
	holotreeVariablesCmd.Flags().BoolVar(&holotreeForce, "force", false, "Force environment creation with refresh.")
	holotreeVariablesCmd.Flags().BoolVar(&holotreeJson, "json", false, "Show environment as JSON.")
}
