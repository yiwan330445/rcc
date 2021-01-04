package main

import (
	"os"

	"github.com/robocorp/rcc/cloud"
	"github.com/robocorp/rcc/cmd"
	"github.com/robocorp/rcc/common"
)

func ExitProtection() {
	status := recover()
	if status != nil {
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

func main() {
	defer os.Stderr.Sync()
	defer os.Stdout.Sync()
	defer ExitProtection()
	cmd.Execute()
}
