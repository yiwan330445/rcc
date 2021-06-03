package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/robocorp/rcc/cloud"
	"github.com/robocorp/rcc/common"
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
		if !config.UsesConda() {
			continue
		}
		condafile := config.CondaConfigFile()
		tree, err := htfs.New()
		pretty.Guard(err == nil, 2, "Holotree creation error: %v", err)
		err = htfs.RecordCondaEnvironment(tree, condafile, false)
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
