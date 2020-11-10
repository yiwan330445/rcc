package cmd

import (
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/operations"

	"github.com/spf13/cobra"
)

var fixCmd = &cobra.Command{
	Use:   "fix",
	Short: "Automatically fix known issues inside robots.",
	Long: `Automatically fix known issues inside robots. Current fixes are:
- make files in PATH folder executable
- convert .sh newlines to unix form`,
	Run: func(cmd *cobra.Command, args []string) {
		if common.Debug {
			defer common.Stopwatch("Fix run lasted").Report()
		}
		err := operations.FixRobot(robotFile)
		if err != nil {
			common.Exit(1, "Error: %v", err)
		}
		common.Log("OK.")
	},
}

func init() {
	robotCmd.AddCommand(fixCmd)
	fixCmd.Flags().StringVarP(&robotFile, "robot", "r", "robot.yaml", "Full path to 'robot.yaml' configuration file.")
}
