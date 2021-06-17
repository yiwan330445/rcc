package cmd

import (
	"github.com/robocorp/rcc/blobs"
	"github.com/robocorp/rcc/pretty"

	"github.com/spf13/cobra"
)

var changelogCmd = &cobra.Command{
	Use:     "changelog",
	Short:   "Show the rcc changelog.",
	Long:    "Show the rcc changelog.",
	Aliases: []string{"changes"},
	Run: func(cmd *cobra.Command, args []string) {
		content, err := blobs.Asset("docs/changelog.md")
		if err != nil {
			pretty.Exit(1, "Cannot show changelog.md, reason: %v", err)
		}
		pretty.Page(content)
	},
}

func init() {
	manCmd.AddCommand(changelogCmd)
}
