package cmd

import (
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/operations"
	"github.com/robocorp/rcc/pretty"
	"github.com/robocorp/rcc/wizard"

	"github.com/spf13/cobra"
)

var wizardConfigCommand = &cobra.Command{
	Use:     "configuration",
	Aliases: []string{"conf", "config", "configure"},
	Short:   "Create a configuration profile for Robocorp tooling interactively.",
	Long:    "Create a configuration profile for Robocorp tooling interactively.",
	Run: func(cmd *cobra.Command, args []string) {
		if !pretty.Interactive {
			pretty.Exit(1, "This is for interactive use only. Do not use in scripting/CI!")
		}
		if common.DebugFlag {
			defer common.Stopwatch("Interactive configuration lasted").Report()
		}
		err := wizard.Configure(args)
		if err != nil {
			pretty.Exit(2, "%v", err)
		}
		_, err = operations.ProduceDiagnostics("", "", false, true)
		if err != nil {
			pretty.Exit(3, "Error: %v", err)
		}
		pretty.Ok()
	},
}

func init() {
	interactiveCmd.AddCommand(wizardConfigCommand)
}
