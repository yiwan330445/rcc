package cmd

import (
	"net/url"
	"path/filepath"
	"strings"

	"github.com/robocorp/rcc/cloud"
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/htfs"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/pretty"
	"github.com/spf13/cobra"
)

var (
	metafileFlag bool
)

func conditionalExpand(filename string) string {
	if !pathlib.IsFile(filename) {
		return filename
	}
	fullpath, err := filepath.Abs(filename)
	if err != nil {
		return filename
	}
	return fullpath
}

func resolveMetafile(link string) ([]string, error) {
	origin, err := url.Parse(link)
	refok := err == nil
	raw, err := cloud.ReadFile(link)
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, line := range strings.SplitAfter(string(raw), "\n") {
		flat := strings.TrimSpace(line)
		if strings.HasPrefix(flat, "#") || len(flat) == 0 {
			continue
		}
		here, err := url.Parse(flat)
		if refok && err == nil {
			relative := origin.ResolveReference(here)
			result = append(result, relative.String())
		} else {
			result = append(result, flat)
		}
	}
	return result, nil
}

func metafileExpansion(links []string, expand bool) []string {
	if !expand {
		return links
	}
	result := []string{}
	for _, metalink := range links {
		links, err := resolveMetafile(conditionalExpand(metalink))
		if err != nil {
			pretty.Warning("Failed to resolve %q metafile, reason: %v", metalink, err)
			continue
		}
		result = append(result, links...)
	}
	return result
}

var holotreePrebuildCmd = &cobra.Command{
	Use:   "prebuild",
	Short: "Prebuild hololib from given set of environment descriptors.",
	Long:  "Prebuild hololib from given set of environment descriptors.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if common.DebugFlag {
			defer common.Stopwatch("Holotree prebuild lasted").Report()
		}

		total, failed := 0, 0

		for _, configfile := range metafileExpansion(args, metafileFlag) {
			total += 1
			pretty.Note("Now building config %q", configfile)
			_, _, err := htfs.NewEnvironment(configfile, "", false, false)
			if err != nil {
				failed += 1
				pretty.Warning("Holotree recording error: %v", err)
			}
		}
		pretty.Guard(failed == 0, 1, "%d out of %d environment builds failed! See output above for details.", failed, total)
		pretty.Ok()
	},
}

func init() {
	holotreeCmd.AddCommand(holotreePrebuildCmd)
	holotreePrebuildCmd.Flags().BoolVarP(&metafileFlag, "metafile", "m", false, "Input arguments are actually files containing links/filenames of environment descriptors.")
}
