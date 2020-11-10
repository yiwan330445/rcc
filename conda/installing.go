package conda

import (
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/shell"
)

func DoInstall() bool {
	if common.Debug {
		defer common.Stopwatch("Installation done in").Report()
	}

	if !ValidateLocations() {
		return false
	}
	install := InstallCommand()
	if common.Debug {
		common.Log("Running: %v", install)
	}
	_, err := shell.New(nil, ".", install...).Transparent()
	if err != nil {
		common.Log("Error: %v", err)
		return false
	}
	return true
}
