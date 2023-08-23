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
	"github.com/robocorp/rcc/operations"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/pretty"

	"github.com/spf13/cobra"
)

var (
	randomIdentifier string
)

func init() {
	randomIdentifier = fmt.Sprintf("%016x", rand.Uint64()^uint64(os.Getpid()))
}

func workingWorm(pipe chan bool, reply chan int, debug bool) {
	if !debug {
		fmt.Fprintf(os.Stderr, "\nWorking: -----")
	}
	seconds := 0
loop:
	for {
		if !debug {
			fmt.Fprintf(os.Stderr, "\b\b\b\b\b%4ds", seconds)
			os.Stderr.Sync()
		}
		select {
		case <-time.After(1 * time.Second):
			seconds += 1
			continue
		case <-pipe:
			break loop
		}
	}
	reply <- seconds
}

var speedtestCmd = &cobra.Command{
	Use:     "speedtest",
	Aliases: []string{"speed"},
	Short:   "Run system speed test to find how rcc performs in your system.",
	Long:    "Run system speed test to find how rcc performs in your system.",
	Run: func(cmd *cobra.Command, args []string) {
		if common.DebugFlag() {
			defer common.Stopwatch("Speed test run lasted").Report()
		}
		common.Log("Running network and filesystem performance tests with %d workers.", anywork.Scale())
		common.Log("This may take several minutes, please be patient.")
		signal := make(chan bool)
		timing := make(chan int)
		silent, debug, trace := common.Silent(), common.DebugFlag(), common.TraceFlag()
		if !debug {
			common.DefineVerbosity(true, false, false)
		}
		go workingWorm(signal, timing, debug)
		folder := common.RobocorpTemp()
		pretty.DebugNote("Speed test will force temporary ROBOCORP_HOME to be %q while testing.", folder)
		err := os.RemoveAll(folder)
		pretty.Guard(err == nil, 4, "Error: %v", err)
		content, err := blobs.Asset("assets/speedtest.yaml")
		pretty.Guard(err == nil, 1, "Error: %v", err)
		condafile := filepath.Join(folder, "speedtest.yaml")
		err = pathlib.WriteFile(condafile, content, 0o666)
		pretty.Guard(err == nil, 2, "Error: %v", err)
		common.ForcedRobocorpHome = folder
		_, score, err := htfs.NewEnvironment(condafile, "", true, true, operations.PullCatalog)
		common.DefineVerbosity(silent, debug, trace)
		pretty.Guard(err == nil, 3, "Error: %v", err)
		common.ForcedRobocorpHome = ""
		err = os.RemoveAll(folder)
		pretty.Guard(err == nil, 4, "Error: %v", err)
		score.Done()
		close(signal)
		elapsed := <-timing
		common.Log("%s", score.Score(anywork.Scale(), elapsed))
		pretty.Ok()
	},
}

func init() {
	configureCmd.AddCommand(speedtestCmd)
}
