package cmd

import (
	"github.com/robocorp/rcc/blobs"
	"github.com/robocorp/rcc/pretty"

	"github.com/spf13/cobra"
)

var manProfilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "Show configuration profiles documentation.",
	Long:  "Show configuration profiles documentation.",
	Run: func(cmd *cobra.Command, args []string) {
		content, err := blobs.Asset("docs/profile_configuration.md")
		if err != nil {
			pretty.Exit(1, "Cannot show profile_configuration.md, reason: %v", err)
		}
		pretty.Page(content)
	},
}

func init() {
	manCmd.AddCommand(manProfilesCmd)
}
