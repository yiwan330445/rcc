package cmd

import (
	"github.com/robocorp/rcc/blobs"
	"github.com/robocorp/rcc/common"

	"github.com/spf13/cobra"
)

var licenseCmd = &cobra.Command{
	Use:   "license",
	Short: "Show the rcc License.",
	Long:  "Show the rcc License.",
	Run: func(cmd *cobra.Command, args []string) {
		content, err := blobs.Asset("assets/man/LICENSE.txt")
		if err != nil {
			common.Exit(1, "Cannot show LICENSE, reason: %v", err)
		}
		common.Log("%s", content)
	},
}

func init() {
	manCmd.AddCommand(licenseCmd)
}
