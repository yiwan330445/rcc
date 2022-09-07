package cmd

import (
	"fmt"
	"os"

	"github.com/robocorp/rcc/pretty"
	"github.com/robocorp/rcc/settings"
	"github.com/spf13/cobra"
)

var settingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Show DEFAULT settings.yaml content. Vanilla rcc settings.",
	Long: `Show DEFAULT settings.yaml content. Vanilla rcc settings.
If you need active status, either use --json option, or "rcc configuration diagnostics".`,
	Run: func(cmd *cobra.Command, args []string) {
		if jsonFlag {
			config, err := settings.SummonSettings()
			pretty.Guard(err == nil, 2, "Error while loading settings: %v", err)
			json, err := config.AsJson()
			pretty.Guard(err == nil, 3, "Error while converting settings: %v", err)
			fmt.Fprintf(os.Stdout, "%s", string(json))
		} else {
			raw, err := settings.DefaultSettings()
			pretty.Guard(err == nil, 1, "Error while loading defaults: %v", err)
			fmt.Fprintf(os.Stdout, "%s", string(raw))
		}
	},
}

func init() {
	configureCmd.AddCommand(settingsCmd)
	settingsCmd.Flags().BoolVarP(&jsonFlag, "json", "j", false, "Show EFFECTIVE settings as JSON stream. For applications to use.")
}
