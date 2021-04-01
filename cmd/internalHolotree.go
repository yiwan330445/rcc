package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime/pprof"

	"github.com/robocorp/rcc/anywork"
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/htfs"
	"github.com/robocorp/rcc/pretty"
	"github.com/spf13/cobra"
)

var (
	holotreeBlueprint []byte
	holotreeProfiled  string
	holotreeSpace     string
	holotreeForce     bool
	holotreeJson      bool
)

var internalHolotreeCmd = &cobra.Command{
	Use:     "holotree conda.yaml+",
	Aliases: []string{"htfs"},
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

		tree, err := htfs.New(common.RobocorpHome())
		pretty.Guard(err == nil, 2, "Failed to create holotree location, reason %v.", err)

		// following must be setup here
		common.StageFolder = tree.Stage()
		common.Stageonly = true
		common.Liveonly = true

		err = os.RemoveAll(tree.Stage())
		pretty.Guard(err == nil, 3, "Failed to clean stage, reason %v.", err)

		profiling := false
		if holotreeProfiled != "" {
			sink, err := os.Create(holotreeProfiled)
			pretty.Guard(err == nil, 5, "Failed to create profile file %q, reason %v.", holotreeProfiled, err)
			defer sink.Close()
			err = pprof.StartCPUProfile(sink)
			pretty.Guard(err == nil, 6, "Failed to start CPU profile, reason %v.", err)
			profiling = true
		}

		common.Debug("Holotree stage is %q.", tree.Stage())
		exists := tree.HasBlueprint(holotreeBlueprint)
		common.Debug("Has blueprint environment: %v", exists)

		if holotreeForce || !exists {
			identityfile := filepath.Join(tree.Stage(), "identity.yaml")
			err = ioutil.WriteFile(identityfile, holotreeBlueprint, 0o640)
			pretty.Guard(err == nil, 3, "Failed to save %q, reason %v.", identityfile, err)
			label, err := conda.NewEnvironment(holotreeForce, identityfile)
			pretty.Guard(err == nil, 3, "Failed to create environment, reason %v.", err)
			common.Debug("Label: %q", label)
		}

		anywork.Scale(17)

		if holotreeForce || !exists {
			err := tree.Record(holotreeBlueprint)
			pretty.Guard(err == nil, 7, "Failed to record blueprint %q, reason: %v", string(holotreeBlueprint), err)
		}

		path, err := tree.Restore(holotreeBlueprint, []byte(common.ControllerIdentity()), []byte(holotreeSpace))
		pretty.Guard(err == nil, 10, "Failed to restore blueprint %q, reason: %v", string(holotreeBlueprint), err)
		if profiling {
			pprof.StopCPUProfile()
		}
		fmt.Fprintln(os.Stderr, path)
		env := conda.EnvironmentExtensionFor(path)
		if holotreeJson {
			asJson(env)
		} else {
			asExportedText(env)
		}
	},
}

func init() {
	internalCmd.AddCommand(internalHolotreeCmd)
	internalHolotreeCmd.Flags().StringVar(&holotreeSpace, "space", "", "Client specific name to identify this environment.")
	internalHolotreeCmd.MarkFlagRequired("space")
	internalHolotreeCmd.Flags().StringVar(&holotreeProfiled, "profile", "", "Filename to save profiling information.")
	internalHolotreeCmd.Flags().BoolVar(&holotreeForce, "force", false, "Force environment creation with refresh.")
	internalHolotreeCmd.Flags().BoolVar(&holotreeJson, "json", false, "Show environment as JSON.")
}
