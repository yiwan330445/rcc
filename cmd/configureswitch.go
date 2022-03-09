package cmd

import (
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/pretty"

	"github.com/spf13/cobra"
)

var configureSwitchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch active configuration profile for Robocorp tooling.",
	Long:  "Switch active configuration profile for Robocorp tooling.",
	Run: func(cmd *cobra.Command, args []string) {
		if common.DebugFlag {
			defer common.Stopwatch("Configuration switch lasted").Report()
		}
		pretty.Ok()
	},
}

func init() {
	configureCmd.AddCommand(configureSwitchCmd)
	configureSwitchCmd.Flags().StringVarP(&profileName, "profile", "p", "", "The name of configuration profile to activate.")
}
