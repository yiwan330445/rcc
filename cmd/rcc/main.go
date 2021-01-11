package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/robocorp/rcc/cloud"
	"github.com/robocorp/rcc/cmd"
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
)

var (
	markedAlready = false
)

func ExitProtection() {
	status := recover()
	if status != nil {
		markTempForRecycling()
		exit, ok := status.(common.ExitCode)
		if ok {
			exit.ShowMessage()
			cloud.WaitTelemetry()
			os.Exit(exit.Code)
		}
		cloud.BackgroundMetric(common.ControllerIdentity(), "rcc.panic.origin", cmd.Origin())
		cloud.WaitTelemetry()
		panic(status)
	}
	cloud.WaitTelemetry()
}

func startTempRecycling() {
	pattern := filepath.Join(conda.RobocorpTempRoot(), "*", "recycle.now")
	found, err := filepath.Glob(pattern)
	if err != nil {
		common.Debug("Recycling failed, reason: %v", err)
		return
	}
	for _, filename := range found {
		go os.RemoveAll(filepath.Dir(filename))
	}
}

func markTempForRecycling() {
	if markedAlready {
		return
	}
	markedAlready = true
	filename := filepath.Join(conda.RobocorpTemp(), "recycle.now")
	ioutil.WriteFile(filename, []byte("True"), 0o640)
	common.Debug("Marked %q for recyling.", conda.RobocorpTemp())
}

func main() {
	go startTempRecycling()
	defer markTempForRecycling()
	defer os.Stderr.Sync()
	defer os.Stdout.Sync()
	defer ExitProtection()
	cmd.Execute()
}
