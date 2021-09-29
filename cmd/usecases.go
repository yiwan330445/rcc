package cmd

import (
	"github.com/robocorp/rcc/blobs"
	"github.com/robocorp/rcc/pretty"

	"github.com/spf13/cobra"
)

var usecasesCmd = &cobra.Command{
	Use:   "usecases",
	Short: "Show some of rcc use cases.",
	Long:  "Show some of rcc use cases.",
	Run: func(cmd *cobra.Command, args []string) {
		content, err := blobs.Asset("docs/usecases.md")
		if err != nil {
			pretty.Exit(1, "Cannot show usecases.md, reason: %v", err)
		}
		pretty.Page(content)
	},
}

func init() {
	manCmd.AddCommand(usecasesCmd)
}
