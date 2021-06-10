package cmd

import (
	"github.com/robocorp/rcc/blobs"
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/pretty"

	"github.com/spf13/cobra"
)

var recipesCmd = &cobra.Command{
	Use:     "recipes",
	Short:   "Show rcc recipes, tips, and tricks.",
	Long:    "Show rcc recipes, tips, and tricks.",
	Aliases: []string{"recipe", "tips", "tricks"},
	Run: func(cmd *cobra.Command, args []string) {
		content, err := blobs.Asset("docs/recipes.md")
		if err != nil {
			pretty.Exit(1, "Cannot show recipes.md, reason: %v", err)
		}
		common.Stdout("\n%s\n", content)
	},
}

func init() {
	manCmd.AddCommand(recipesCmd)
}
