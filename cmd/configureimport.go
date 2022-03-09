package cmd

import (
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/pretty"

	"github.com/spf13/cobra"
)

var configureImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import a configuration profile for Robocorp tooling.",
	Long:  "Import a configuration profile for Robocorp tooling.",
	Run: func(cmd *cobra.Command, args []string) {
		if common.DebugFlag {
			defer common.Stopwatch("Configuration import lasted").Report()
		}
		pretty.Ok()
	},
}

func init() {
	configureCmd.AddCommand(configureImportCmd)
	configureImportCmd.Flags().StringVarP(&configFile, "filename", "f", "local_config.yaml", "The filename to import as configuration profile.")
}
