package cmd

import (
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/operations"
	"github.com/robocorp/rcc/pretty"

	"github.com/spf13/cobra"
)

var (
	fileOption  string
	robotOption string
)

var diagnosticsCmd = &cobra.Command{
	Use:   "diagnostics",
	Short: "Run system diagnostics to help resolve rcc issues.",
	Long:  "Run system diagnostics to help resolve rcc issues.",
	Run: func(cmd *cobra.Command, args []string) {
		if common.DebugFlag {
			defer common.Stopwatch("Diagnostic run lasted").Report()
		}
		_, err := operations.ProduceDiagnostics(fileOption, robotOption, jsonFlag)
		if err != nil {
			pretty.Exit(1, "Error: %v", err)
		}
		pretty.Ok()
	},
}

func init() {
	configureCmd.AddCommand(diagnosticsCmd)
	diagnosticsCmd.Flags().BoolVarP(&jsonFlag, "json", "j", false, "Output in JSON format")
	diagnosticsCmd.Flags().StringVarP(&fileOption, "file", "f", "", "Save output into a file.")
	diagnosticsCmd.Flags().StringVarP(&robotOption, "robot", "r", "", "Full path to 'robot.yaml' configuration file. [optional]")
}
