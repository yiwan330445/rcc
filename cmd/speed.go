package cmd

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/robocorp/rcc/anywork"
	"github.com/robocorp/rcc/blobs"
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/htfs"
	"github.com/robocorp/rcc/pretty"

	"github.com/spf13/cobra"
)

var (
	randomIdentifier string
)

func init() {
	randomIdentifier = fmt.Sprintf("%016x", rand.Uint64()^uint64(os.Getpid()))
}

func workingWorm(pipe chan bool) {
	fmt.Fprintf(os.Stderr, "\nWorking: -----")
	index := 0
loop:
	for {
		fmt.Fprintf(os.Stderr, "\b\b\b\b\b%4ds", index)
		os.Stderr.Sync()
		select {
		case <-time.After(1 * time.Second):
			index += 1
			continue
		case <-pipe:
			break loop
		}
	}
}

var speedtestCmd = &cobra.Command{
	Use:     "speedtest",
	Aliases: []string{"speed"},
	Short:   "Run system speed test to find how rcc performs in your system.",
	Long:    "Run system speed test to find how rcc performs in your system.",
	Run: func(cmd *cobra.Command, args []string) {
		if common.DebugFlag {
			defer common.Stopwatch("Speed test run lasted").Report()
		}
		common.Log("Running network and filesystem performance tests with %d workers.", anywork.Scale())
		common.Log("This may take several minutes, please be patient.")
		signal := make(chan bool)
		go workingWorm(signal)
		silent, trace, debug := common.Silent, common.TraceFlag, common.DebugFlag
		common.Silent = true
		common.UnifyVerbosityFlags()
		folder := common.RobocorpTemp()
		content, err := blobs.Asset("assets/speedtest.yaml")
		if err != nil {
			pretty.Exit(1, "Error: %v", err)
		}
		condafile := filepath.Join(folder, "speedtest.yaml")
		err = os.WriteFile(condafile, content, 0o666)
		if err != nil {
			pretty.Exit(2, "Error: %v", err)
		}
		common.ForcedRobocorpHome = folder
		_, score, err := htfs.NewEnvironment(condafile, "", true, true)
		common.Silent, common.TraceFlag, common.DebugFlag = silent, trace, debug
		common.UnifyVerbosityFlags()
		if err != nil {
			pretty.Exit(3, "Error: %v", err)
		}
		common.ForcedRobocorpHome = ""
		err = os.RemoveAll(folder)
		if err != nil {
			pretty.Exit(4, "Error: %v", err)
		}
		score.Done()
		close(signal)
		common.Log("%s", score.Score())
		pretty.Ok()
	},
}

func init() {
	configureCmd.AddCommand(speedtestCmd)
}
