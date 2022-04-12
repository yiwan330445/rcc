package cmd

import (
	"os"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/pretty"

	"github.com/spf13/cobra"
)

var holotreeInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize shared holotree location.",
	Long:  "Initialize shared holotree location.",
	Run: func(cmd *cobra.Command, args []string) {
		if common.DebugFlag {
			defer common.Stopwatch("Conda YAML hash calculation lasted").Report()
		}
		pretty.Guard(common.FixedHolotreeLocation(), 1, "Fixed Holotree is not available in this system!")
		if os.Geteuid() > 0 {
			pretty.Warning("Running this command might need sudo/root access rights. Still, trying ...")
		}
		_, err := pathlib.MakeSharedDir(common.HoloLocation())
		pretty.Guard(err == nil, 2, "Could not enable shared location at %q, reason: %v", common.HoloLocation(), err)
		pretty.Ok()
	},
}

func init() {
	holotreeCmd.AddCommand(holotreeInitCmd)
}
