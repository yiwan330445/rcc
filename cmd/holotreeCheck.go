package cmd

import (
	"fmt"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/htfs"
	"github.com/robocorp/rcc/pretty"
	"github.com/spf13/cobra"
)

var holotreeCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check holotree library integrity.",
	Long:  "Check holotree library integrity.",
	Run: func(cmd *cobra.Command, args []string) {
		fs, err := htfs.NewRoot(common.HololibLibraryLocation())
		pretty.Guard(err == nil, 1, "%s", err)
		err = fs.Lift()
		pretty.Guard(err == nil, 2, "%s", err)
		err = fs.AllFiles(htfs.Hasher())
		pretty.Guard(err == nil, 3, "%s", err)
		collector := make(map[string]string)
		err = fs.Treetop(htfs.IntegrityCheck(collector))
		pretty.Guard(err == nil, 4, "%s", err)
		for k, v := range collector {
			fmt.Println(k, v)
		}
		pretty.Guard(len(collector) == 0, 5, "Size: %d", len(collector))
		pretty.Ok()
	},
}

func init() {
	holotreeCmd.AddCommand(holotreeCheckCmd)
}
