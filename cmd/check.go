package cmd

import (
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:     "check",
	Aliases: []string{"c"},
	Short:   "Check if conda is installed in managed location.",
	Long: `Check if conda is installed. And optionally also force download and install
conda using "rcc conda download" and "rcc conda install" commands.  `,
	Run: func(cmd *cobra.Command, args []string) {
		if common.Debug {
			defer common.Stopwatch("Conda check took").Report()
		}
		if conda.HasConda() {
			common.Exit(0, "OK.")
		}
		if common.Debug {
			common.Log("Conda is missing ...")
		}
		if !autoInstall {
			common.Exit(1, "Error: No conda.")
		}
		if common.Debug {
			common.Log("Starting conda download ...")
		}
		if !conda.DoDownload() {
			common.Exit(2, "Error: Conda download failed.")
		}
		if common.Debug {
			common.Log("Starting conda install ...")
		}
		if !conda.DoInstall() {
			common.Exit(3, "Error: Conda install failed.")
		}
		if common.Debug {
			common.Log("Conda install completed ...")
		}
		common.Log("OK.")
	},
}

func init() {
	condaCmd.AddCommand(checkCmd)

	checkCmd.Flags().BoolVarP(&autoInstall, "install", "i", false, "If conda is missing, download and install it automatically.")
}
