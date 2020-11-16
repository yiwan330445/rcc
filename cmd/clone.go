package cmd

import (
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/pretty"

	"github.com/spf13/cobra"
)

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Debug tool for cloning folders.",
	Long:  `Internal debug tool for checking cloning speed in various disk drives.`,
	Run: func(cmd *cobra.Command, args []string) {
		source := cmd.LocalFlags().Lookup("source").Value.String()
		target := cmd.LocalFlags().Lookup("target").Value.String()
		defer common.Stopwatch("rcc internal clone lasted").Report()
		success := conda.CloneFromTo(source, target)
		common.Log("Was successful: %v", success)
		if !success {
			pretty.Exit(1, "Error: Cloning failed.")
		}
	},
}

func init() {
	internalCmd.AddCommand(cloneCmd)

	cloneCmd.Flags().StringP("source", "s", "", "Source directory to clone.")
	cloneCmd.Flags().StringP("target", "t", "", "Source directory to clone.")
	cloneCmd.MarkFlagRequired("source")
	cloneCmd.MarkFlagRequired("target")
}
