package cmd

import (
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/pretty"

	"github.com/spf13/cobra"
)

var (
	allFlag        bool
	orphanFlag     bool
	minicondaFlag  bool
	micromambaFlag bool
	daysOption     int
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Cleanup old managed virtual environments.",
	Long: `Cleanup removes old virtual environments from existence.
After cleanup, they will not be available anymore.`,
	Run: func(cmd *cobra.Command, args []string) {
		if common.DebugFlag {
			defer common.Stopwatch("Env cleanup lasted").Report()
		}
		err := conda.Cleanup(daysOption, dryFlag, orphanFlag, allFlag, minicondaFlag, micromambaFlag)
		if err != nil {
			pretty.Exit(1, "Error: %v", err)
		}
		pretty.Ok()
	},
}

func init() {
	envCmd.AddCommand(cleanupCmd)
	cleanupCmd.Flags().BoolVarP(&dryFlag, "dryrun", "d", false, "Don't delete environments, just show what would happen.")
	cleanupCmd.Flags().BoolVarP(&orphanFlag, "orphans", "o", false, "Cleanup orphan, unreachable enviroments.")
	cleanupCmd.Flags().BoolVarP(&minicondaFlag, "miniconda", "m", false, "Remove miniconda3 installation (replaced by micromamba).")
	cleanupCmd.Flags().BoolVarP(&micromambaFlag, "micromamba", "", false, "Remove micromamba installation.")
	cleanupCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Cleanup all enviroments.")
	cleanupCmd.Flags().IntVarP(&daysOption, "days", "", 30, "What is the limit in days to keep environments for (deletes environments older than this).")
}
