package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/robocorp/rcc/cloud"
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/htfs"
	"github.com/robocorp/rcc/operations"
	"github.com/robocorp/rcc/pretty"
	"github.com/robocorp/rcc/robot"
	"github.com/robocorp/rcc/settings"
	"github.com/spf13/cobra"
)

var (
	holotreeQuick bool
)

func updateEnvironments(robots []string) {
	for at, robotling := range robots {
		workarea := filepath.Join(os.TempDir(), fmt.Sprintf("workarea%x%x", common.When, at))
		defer os.RemoveAll(workarea)
		common.Debug("Using temporary workarea: %v", workarea)
		err := operations.Unzip(workarea, robotling, false, true)
		pretty.Guard(err == nil, 2, "Could not unzip %q, reason: %w", robotling, err)
		targetRobot := robot.DetectConfigurationName(workarea)
		config, err := robot.LoadRobotYaml(targetRobot, false)
		pretty.Guard(err == nil, 2, "Could not load robot config %q, reason: %w", targetRobot, err)
		condafile := config.CondaConfigFile()
		right, err := conda.ReadCondaYaml(condafile)
		pretty.Guard(err == nil, 2, "Could not load environmet config %q, reason: %w", condafile, err)
		content, err := right.AsYaml()
		pretty.Guard(err == nil, 2, "YAML error: %v", err)
		holotreeBlueprint := []byte(content)
		_, err = htfs.RecordEnvironment(holotreeBlueprint, false)
		pretty.Guard(err == nil, 2, "Holotree recording error: %v", err)
	}
}

var holotreeBootstrapCmd = &cobra.Command{
	Use:     "bootstrap",
	Aliases: []string{"boot"},
	Short:   "Bootstrap holotree from set of templates.",
	Long:    "Bootstrap holotree from set of templates.",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if common.DebugFlag {
			defer common.Stopwatch("Holotree bootstrap lasted").Report()
		}

		ok := conda.MustMicromamba()
		pretty.Guard(ok, 1, "Could not get micromamba installed.")

		robots := make([]string, 0, 20)
		for key, _ := range settings.Global.Templates() {
			zipname := fmt.Sprintf("%s.zip", key)
			filename := filepath.Join(common.TemplateLocation(), zipname)
			robots = append(robots, filename)
			url := fmt.Sprintf("templates/%s", zipname)
			err := cloud.Download(settings.Global.DownloadsLink(url), filename)
			pretty.Guard(err == nil, 2, "Could not download %q, reason: %w", url, err)
		}

		if !holotreeQuick {
			updateEnvironments(robots)
		}

		pretty.Ok()
	},
}

func init() {
	holotreeCmd.AddCommand(holotreeBootstrapCmd)
	holotreeBootstrapCmd.Flags().BoolVar(&holotreeQuick, "quick", false, "Do not create environments, just download templates.")
}
