package cmd

import (
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/operations"
	"github.com/robocorp/rcc/pretty"
	"github.com/spf13/cobra"
)

var holotreeImportCmd = &cobra.Command{
	Use:   "import hololib.zip+",
	Short: "Import one or more hololib.zip files into local hololib.",
	Long:  "Import one or more hololib.zip files into local hololib.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if common.DebugFlag {
			defer common.Stopwatch("Holotree import command lasted").Report()
		}
		for _, filename := range args {
			err := operations.Unzip(common.HololibLocation(), filename, true, false)
			pretty.Guard(err == nil, 1, "Could not import %q, reason: %v", filename, err)
		}
		pretty.Ok()
	},
}

func init() {
	holotreeCmd.AddCommand(holotreeImportCmd)
}
