package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/robocorp/rcc/htfs"
	"github.com/robocorp/rcc/pretty"

	"github.com/spf13/cobra"
)

var holotreePlanCmd = &cobra.Command{
	Use:   "plan",
	Short: "Show installation plans for given holotree spaces (or substrings)",
	Long:  "Show installation plans for given holotree spaces (or substrings)",

	Run: func(cmd *cobra.Command, args []string) {
		for _, prefix := range args {
			for _, label := range htfs.FindEnvironment(prefix) {
				planfile, ok := htfs.InstallationPlan(label)
				pretty.Guard(ok, 1, "Could not find plan for: %v", label)
				content, err := ioutil.ReadFile(planfile)
				pretty.Guard(err == nil, 2, "Could not read plan %q, reason: %v", planfile, err)
				fmt.Fprintf(os.Stdout, string(content))
			}
		}
	},
}

func init() {
	holotreeCmd.AddCommand(holotreePlanCmd)
}
