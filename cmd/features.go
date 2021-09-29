package cmd

import (
	"github.com/robocorp/rcc/blobs"
	"github.com/robocorp/rcc/pretty"

	"github.com/spf13/cobra"
)

var featuresCmd = &cobra.Command{
	Use:   "features",
	Short: "Show some of rcc features.",
	Long:  "Show some of rcc features.",
	Run: func(cmd *cobra.Command, args []string) {
		content, err := blobs.Asset("docs/features.md")
		if err != nil {
			pretty.Exit(1, "Cannot show features.md, reason: %v", err)
		}
		pretty.Page(content)
	},
}

func init() {
	manCmd.AddCommand(featuresCmd)
}
