package cmd

import (
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/operations"

	"github.com/spf13/cobra"
)

var wrapCmd = &cobra.Command{
	Use:   "wrap",
	Short: "Build a robot out of directory content.",
	Long: `Build a robot out of directory content. This command expects to get robot
filename, source directory and optional ignore files. When wrap is run again
existing robot file will silently be overwritten..`,
	Run: func(cmd *cobra.Command, args []string) {
		if common.Debug {
			defer common.Stopwatch("Wrap lasted").Report()
		}
		err := operations.Zip(directory, zipfile, ignores)
		if err != nil {
			common.Exit(1, "Error: %v", err)
		}
		common.Log("OK.")
	},
}

func init() {
	robotCmd.AddCommand(wrapCmd)
	wrapCmd.Flags().StringVarP(&zipfile, "zipfile", "z", "robot.zip", "The filename for the robot.")
	wrapCmd.Flags().StringVarP(&directory, "directory", "d", ".", "The root directory create the robot from.")
	wrapCmd.Flags().StringArrayVarP(&ignores, "ignore", "i", []string{}, "File with ignore patterns.")
}
