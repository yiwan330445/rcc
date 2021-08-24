package cmd

import (
	"fmt"

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
	for k, v := range collector {
		fmt.Println(k, v)
	}
	pretty.Guard(len(collector) == 0, 5, "Size: %d", len(collector))
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
