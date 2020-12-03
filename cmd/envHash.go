package cmd

import (
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/pretty"

	"github.com/spf13/cobra"
)

var envHashCmd = &cobra.Command{
	Use:   "hash <conda.yaml+>",
	Short: "Calculates a hash for managed virtual environment from conda.yaml files.",
	Long:  "Calculates a hash for managed virtual environment from conda.yaml files.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if common.DebugFlag {
			defer common.Stopwatch("Conda YAML hash calculation lasted").Report()
		}
		hash, err := conda.CalculateComboHash(args...)
		if err != nil {
			pretty.Exit(1, "Hash calculation failed: %v", err)
		} else {
			common.Log("Hash for %v is %v.", args, hash)
		}
		if common.Silent {
			common.Stdout("%s\n", hash)
		}
	},
}

func init() {
	envCmd.AddCommand(envHashCmd)
}
