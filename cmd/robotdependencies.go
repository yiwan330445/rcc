package cmd

import (
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/operations"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/pretty"
	"github.com/robocorp/rcc/robot"

	"github.com/spf13/cobra"
)

var (
	copyDependenciesFlag bool
)

func doShowDependencies(config robot.Robot, label string) {
	filename, _ := config.DependenciesFile()
	err := conda.SideBySideViewOfDependencies(conda.GoldenMasterFilename(label), filename)
	pretty.Guard(err == nil, 3, "Failed to show dependencies, reason: %v", err)
}

func doCopyDependencies(config robot.Robot, label string) {
	mode := "[create]"
	target, found := config.DependenciesFile()
	if found {
		mode = "[overwrite]"
	}
	source := conda.GoldenMasterFilename(label)
	common.Log("%sCopying %q as wanted %q %s.%s", pretty.Yellow, source, target, mode, pretty.Reset)
	err := pathlib.CopyFile(source, target, found)
	pretty.Guard(err == nil, 2, "Copy %q -> %q failed, reason: %v", source, target, err)
}

func doShowIdeal(config robot.Robot, label string) {
	ideal, ok := config.IdealCondaYaml()
	pretty.Guard(ok, 4, "Could not determine ideal conda.yaml. Sorry.")
	common.Log("Ideal conda.yaml based on 'dependencies.yaml' would be:\n%s", ideal)
}

var robotDependenciesCmd = &cobra.Command{
	Use:     "dependencies",
	Short:   "View wanted vs. available dependencies of robot execution environment.",
	Long:    "View wanted vs. available dependencies of robot execution environment.",
	Aliases: []string{"deps"},
	Run: func(cmd *cobra.Command, args []string) {
		if common.DebugFlag {
			defer common.Stopwatch("Robot dependencies run lasted").Report()
		}
		simple, config, _, label := operations.LoadAnyTaskEnvironment(robotFile, forceFlag)
		pretty.Guard(!simple, 1, "Cannot view dependencies of simple robots.")
		if copyDependenciesFlag {
			common.Log("--")
			doCopyDependencies(config, label)
		}
		common.Log("--")
		doShowDependencies(config, label)
		doShowIdeal(config, label)
		pretty.Ok()
	},
}

func init() {
	robotCmd.AddCommand(robotDependenciesCmd)
	robotDependenciesCmd.Flags().BoolVarP(&copyDependenciesFlag, "copy", "c", false, "Copy golden-ee.yaml from environment as wanted dependencies.yaml, overwriting previous if exists.")
	robotDependenciesCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Forced environment update.")
	robotDependenciesCmd.Flags().StringVarP(&robotFile, "robot", "r", "robot.yaml", "Full path to the 'robot.yaml' configuration file.")
	robotDependenciesCmd.Flags().StringVarP(&common.HolotreeSpace, "space", "s", "", "Space to use for execution environment dependencies.")
}
