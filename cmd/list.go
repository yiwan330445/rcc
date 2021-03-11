package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/pretty"

	"github.com/spf13/cobra"
)

func textDump(lines []string) {
	sort.Strings(lines)
	for _, line := range lines {
		common.Log("%s", line)
	}
}

func jsonDump(entries map[string]interface{}) {
	body, err := json.MarshalIndent(entries, "", "  ")
	if err == nil {
		fmt.Fprintln(os.Stdout, string(body))
	}
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Listing currently managed virtual environments.",
	Long: `List shows listing of currently managed virtual environments
in human readable form.`,
	Run: func(cmd *cobra.Command, args []string) {
		templates := conda.TemplateList()
		if len(templates) == 0 {
			pretty.Exit(1, "No environments available.")
		}
		lines := make([]string, 0, len(templates))
		entries := make(map[string]interface{})
		if !jsonFlag {
			common.Log("%-25s  %-25s  %-16s  %5s  %s", "Last used", "Last cloned", "Environment", "Plan?", "Leased duration")
		}
		for _, template := range templates {
			details := make(map[string]interface{})
			entries[template] = details
			cloned := "N/A"
			used := cloned
			when, err := conda.LastUsed(conda.TemplateFrom(template))
			if err == nil {
				cloned = when.Format(time.RFC3339)
			}
			when, err = conda.LastUsed(conda.LiveFrom(template))
			if err == nil {
				used = when.Format(time.RFC3339)
			}
			details["name"] = template
			details["used"] = used
			details["cloned"] = cloned
			details["leased"] = conda.WhoLeased(template)
			details["expires"] = conda.LeaseExpires(template)
			details["base"] = conda.TemplateFrom(template)
			details["live"] = conda.LiveFrom(template)
			planfile, plan := conda.InstallationPlan(template)
			lines = append(lines, fmt.Sprintf("%-25s  %-25s  %-16s  %5v  %q %s", used, cloned, template, plan, conda.WhoLeased(template), conda.LeaseExpires(template)))
			if plan {
				details["plan"] = planfile
			}
		}
		if jsonFlag {
			jsonDump(entries)
		} else {
			textDump(lines)
		}
	},
}

func init() {
	envCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVarP(&jsonFlag, "json", "j", false, "Output in JSON format.")
}
