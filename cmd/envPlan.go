package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/pretty"

	"github.com/spf13/cobra"
)

var envPlanCmd = &cobra.Command{
	Use:   "plan",
	Short: "Show installation plan for given environment (or prefix)",
	Long:  "Show installation plan for given environment (or prefix)",

	Run: func(cmd *cobra.Command, args []string) {
		for _, prefix := range args {
			for _, label := range conda.FindEnvironment(prefix) {
				planfile, ok := conda.InstallationPlan(label)
				pretty.Guard(ok, 1, "Could not find plan for: %v", label)
				content, err := ioutil.ReadFile(planfile)
				pretty.Guard(err == nil, 2, "Could not read plan %q, reason: %v", planfile, err)
				fmt.Fprintf(os.Stdout, string(content))
			}
		}
	},
}

func init() {
	envCmd.AddCommand(envPlanCmd)
}
