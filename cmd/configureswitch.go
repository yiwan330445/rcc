package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/operations"
	"github.com/robocorp/rcc/pretty"
	"github.com/robocorp/rcc/settings"

	"github.com/spf13/cobra"
)

func profileMap() map[string]string {
	pattern := common.ExpandPath(filepath.Join(common.RobocorpHome(), "profile_*.yaml"))
	found, err := filepath.Glob(pattern)
	pretty.Guard(err == nil, 1, "Error while searching profiles: %v", err)
	result := make(map[string]string)
	for _, name := range found {
		profile := settings.Profile{}
		err = profile.LoadFrom(name)
		if err == nil {
			result[profile.Name] = profile.Description
		}
	}
	return result
}

func jsonListProfiles() {
	content, err := operations.NiceJsonOutput(profileMap())
	pretty.Guard(err == nil, 1, "Error serializing profiles: %v", err)
	common.Stdout("%s\n", content)
}

func listProfiles() {
	profiles := profileMap()
	pretty.Guard(len(profiles) > 0, 2, "No profiles found, you must first import some.")
	common.Stdout("Available profiles:\n")
	for name, description := range profiles {
		common.Stdout("- %s: %s\n", name, description)
	}
	common.Stdout("\n")
}

func switchProfileTo(name string) {
	filename := fmt.Sprintf("profile_%s.yaml", strings.ToLower(name))
	fullpath := common.ExpandPath(filepath.Join(common.RobocorpHome(), filename))
	profile := settings.Profile{}
	err := profile.LoadFrom(fullpath)
	pretty.Guard(err == nil, 3, "Error while loading/parsing profile, reason: %v", err)
	err = profile.Activate()
	pretty.Guard(err == nil, 4, "Error while activating profile, reason: %v", err)
}

func cleanupProfile() {
	profile := settings.Profile{}
	err := profile.Remove()
	pretty.Guard(err == nil, 5, "Error while clearing profile, reason: %v", err)
}

var configureSwitchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch active configuration profile for Robocorp tooling.",
	Long:  "Switch active configuration profile for Robocorp tooling.",
	Run: func(cmd *cobra.Command, args []string) {
		if common.DebugFlag {
			defer common.Stopwatch("Configuration switch lasted").Report()
		}
		if clearProfile {
			cleanupProfile()
			pretty.Ok()
		} else if len(profileName) == 0 {
			if jsonFlag {
				jsonListProfiles()
			} else {
				listProfiles()
				common.Stdout("Currently active profile is: %s\n", settings.Global.Name())
				pretty.Ok()
			}
		} else {
			switchProfileTo(profileName)
			pretty.Ok()
		}
	},
}

func init() {
	configureCmd.AddCommand(configureSwitchCmd)
	configureSwitchCmd.Flags().StringVarP(&profileName, "profile", "p", "", "The name of configuration profile to activate.")
	configureSwitchCmd.Flags().BoolVarP(&clearProfile, "noprofile", "n", false, "Remove active profile, and reset to defaults.")
	configureSwitchCmd.Flags().BoolVarP(&jsonFlag, "json", "j", false, "Show profile list as JSON stream.")
}
