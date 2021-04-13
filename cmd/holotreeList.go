package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/htfs"
	"github.com/robocorp/rcc/pretty"
	"github.com/spf13/cobra"
)

func humaneHolotreeSpaceListing(tree htfs.Library) {
	tabbed := tabwriter.NewWriter(os.Stderr, 2, 4, 2, ' ', 0)
	tabbed.Write([]byte("Identity\tController\tSpace\tBlueprint\tFull path\n"))
	tabbed.Write([]byte("--------\t----------\t-----\t--------\t---------\n"))
	for _, space := range tree.Spaces() {
		data := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\n", space.Identity, space.Controller, space.Space, space.Blueprint, space.Path)
		tabbed.Write([]byte(data))
	}
	tabbed.Flush()
}

func jsonicHolotreeSpaceListing(tree htfs.Library) {
	details := make(map[string]map[string]string)
	for _, space := range tree.Spaces() {
		hold, ok := details[space.Identity]
		if !ok {
			hold = make(map[string]string)
			details[space.Identity] = hold
			hold["id"] = space.Identity
			hold["controller"] = space.Controller
			hold["space"] = space.Space
			hold["blueprint"] = space.Blueprint
			hold["path"] = space.Path
			hold["meta"] = space.Path + ".meta"
			hold["spec"] = filepath.Join(space.Path, "identity.yaml")
			hold["plan"] = filepath.Join(space.Path, "rcc_plan.log")
		}
	}
	body, err := json.MarshalIndent(details, "", "  ")
	pretty.Guard(err == nil, 1, "Could not create json, reason: %w", err)
	fmt.Println(string(body))
}

var holotreeListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List holotree spaces.",
	Long:    "List holotree spaces.",
	Run: func(cmd *cobra.Command, args []string) {
		if common.DebugFlag {
			defer common.Stopwatch("Holotree list lasted").Report()
		}

		tree, err := htfs.New(common.RobocorpHome())
		pretty.Guard(err == nil, 1, "Could not get holotree, reason: %w", err)

		if jsonFlag {
			jsonicHolotreeSpaceListing(tree)
		} else {
			humaneHolotreeSpaceListing(tree)
		}

	},
}

func init() {
	holotreeCmd.AddCommand(holotreeListCmd)
	holotreeListCmd.Flags().BoolVarP(&jsonFlag, "json", "j", false, "Output in JSON format")
}
