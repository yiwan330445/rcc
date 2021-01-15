package cmd

import (
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/pretty"

	"github.com/spf13/cobra"
)

var (
	leaseHash string
)

var envUnleaseCmd = &cobra.Command{
	Use:   "unlease",
	Short: "Drop existing lease of given environment.",
	Long:  "Drop existing lease of given environment.",
	Run: func(cmd *cobra.Command, args []string) {
		if common.DebugFlag {
			defer common.Stopwatch("Conda env unlease lasted").Report()
		}
		err := conda.DropLease(leaseHash, common.LeaseContract)
		if err != nil {
			pretty.Exit(1, "Error: %v", err)
		}
		pretty.Ok()
	},
}

func init() {
	envCmd.AddCommand(envUnleaseCmd)
	envUnleaseCmd.Flags().StringVar(&common.LeaseContract, "lease", "", "unique lease contract for long living environments")
	envUnleaseCmd.MarkFlagRequired("lease")
	envUnleaseCmd.Flags().StringVar(&leaseHash, "hash", "", "hash identity of leased environment")
	envUnleaseCmd.MarkFlagRequired("hash")
}
