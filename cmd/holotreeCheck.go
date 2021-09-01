package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/robocorp/rcc/anywork"
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/htfs"
	"github.com/robocorp/rcc/pretty"
	"github.com/spf13/cobra"
)

func checkHolotreeIntegrity() {
	common.Timeline("holotree integrity check start")
	defer common.Timeline("holotree integrity check done")
	fs, err := htfs.NewRoot(common.HololibLibraryLocation())
	pretty.Guard(err == nil, 1, "%s", err)
	common.Timeline("holotree integrity lift")
	err = fs.Lift()
	pretty.Guard(err == nil, 2, "%s", err)
	common.Timeline("holotree integrity hasher")
	known := htfs.LoadHololibHashes()
	err = fs.AllFiles(htfs.Hasher(known))
	pretty.Guard(err == nil, 3, "%s", err)
	collector := make(map[string]string)
	common.Timeline("holotree integrity collector")
	err = fs.Treetop(htfs.IntegrityCheck(collector))
	common.Timeline("holotree integrity report")
	pretty.Guard(err == nil, 4, "%s", err)
	purge := make(map[string]bool)
	for k, v := range collector {
		fmt.Println(k, v)
		found, ok := known[filepath.Base(k)]
		if !ok {
			continue
		}
		for catalog, _ := range found {
			purge[catalog] = true
		}
	}
	redo := false
	for k, _ := range purge {
		fmt.Println("Purge catalog:", k)
		redo = true
		anywork.Backlog(htfs.RemoveFile(k))
	}
	if redo {
		pretty.Warning("Some catalogs were purged. Run this check command again, please!")
	}
	err = anywork.Sync()
	pretty.Guard(err == nil, 5, "%s", err)
	pretty.Guard(len(collector) == 0, 6, "Size: %d", len(collector))
}

var holotreeCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check holotree library integrity.",
	Long:  "Check holotree library integrity.",
	Run: func(cmd *cobra.Command, args []string) {
		checkHolotreeIntegrity()
		pretty.Ok()
	},
}

func init() {
	holotreeCmd.AddCommand(holotreeCheckCmd)
}
