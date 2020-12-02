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
			os.Exit(exit.Code)
		}
		cloud.SendMetric(common.ControllerIdentity(), "rcc.panic.origin", cmd.Origin())
		panic(status)
	}
}

func main() {
	defer os.Stderr.Sync()
	defer os.Stdout.Sync()
	defer ExitProtection()
	cmd.Execute()
}
