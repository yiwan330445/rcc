package cmd

import (
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"

	"github.com/spf13/cobra"
)

func doDownload() bool {
	if common.Debug {
		defer common.Stopwatch("Download done in").Report()
	}

	err := conda.DownloadConda()
	if err != nil {
		common.Log("FAILURE: %s", err)
		return false
	} else {
		common.Log("Verify checksum from https://docs.conda.io/en/latest/miniconda.html")
		return true
	}
}

var condaDownloadCmd = &cobra.Command{
	Use:     "download",
	Aliases: []string{"dl", "d"},
	Short:   "Download the miniconda3 installer.",
	Long:    `Downloads the miniconda3 installer for this platform.`,
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if !doDownload() {
			common.Exit(1, "Download failed.")
		}
	},
}

func init() {
	condaCmd.AddCommand(condaDownloadCmd)
}
