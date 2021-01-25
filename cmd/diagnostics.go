package cmd

import (
	"fmt"
	"sort"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/operations"
	"github.com/robocorp/rcc/pretty"

	"github.com/spf13/cobra"
)

func jsonDiagnostics(details *operations.DiagnosticStatus) {
	form, err := details.AsJson()
	if err != nil {
		pretty.Exit(1, "Error: %s", err)
	}
	fmt.Println(form)
}

func humaneDiagnostics(details *operations.DiagnosticStatus) {
	common.Log("Diagnostics:")
	keys := make([]string, 0, len(details.Details))
	for key, _ := range details.Details {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := details.Details[key]
		common.Log(" - %-18s...  %q", key, value)
	}
	common.Log("")
	common.Log("Checks:")
	for _, check := range details.Checks {
		common.Log(" - %-8s %-8s %s", check.Type, check.Status, check.Message)
	}
}

var diagnosticsCmd = &cobra.Command{
	Use:   "diagnostics",
	Short: "Run system diagnostics to help resolve rcc issues.",
	Long:  "Run system diagnostics to help resolve rcc issues.",
	Run: func(cmd *cobra.Command, args []string) {
		if common.DebugFlag {
			defer common.Stopwatch("Diagnostic run lasted").Report()
		}
		result := operations.RunDiagnostics()
		if jsonFlag {
			jsonDiagnostics(result)
		} else {
			humaneDiagnostics(result)
			pretty.Ok()
		}
	},
}

func init() {
	configureCmd.AddCommand(diagnosticsCmd)
	diagnosticsCmd.Flags().BoolVarP(&jsonFlag, "json", "j", false, "Output in JSON format")
}
