package cmd

import (
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/pretty"
	"github.com/robocorp/rcc/wizard"

	"github.com/spf13/cobra"
)

var altFlag bool

var wizardCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a directory structure for a robot interactively.",
	Long:  "Create a directory structure for a robot interactively.",
	Run: func(cmd *cobra.Command, args []string) {
		if !pretty.Interactive {
			pretty.Exit(1, "This is for interactive use only. Do not use in scripting/CI!")
		}
		if common.DebugFlag {
			defer common.Stopwatch("Interactive create lasted").Report()
		}
		if altFlag {
			err := wizard.AltCreate(args)
			if err != nil {
				pretty.Exit(2, "%v", err)
			}
		} else {
			err := wizard.Create(args)
			if err != nil {
				pretty.Exit(2, "%v", err)
			}
		}
	},
}

func init() {
	interactiveCmd.AddCommand(wizardCreateCmd)
	rootCmd.AddCommand(wizardCreateCmd)
	wizardCreateCmd.Flags().BoolVarP(&altFlag, "alt", "a", false, "select alternative create command")
}
