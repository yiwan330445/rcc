package conda

import (
	"os"

	"github.com/robocorp/rcc/cloud"
	"github.com/robocorp/rcc/common"
)

func MustMicromamba() bool {
	return HasMicroMamba() || ((DoDownload() || DoDownload() || DoDownload()) && DoInstall())
}

func DoDownload() bool {
	if common.DebugFlag {
		defer common.Stopwatch("Download done in").Report()
	}

	common.Log("Downloading micromamba, this may take awhile ...")

	err := DownloadMicromamba()
	if err != nil {
		common.Error("Download", err)
		os.Remove(BinMicromamba())
		return false
	}
	cloud.BackgroundMetric(common.ControllerIdentity(), "rcc.micromamba.download", common.Version)
	return true
}

func DoInstall() bool {
	if common.DebugFlag {
		defer common.Stopwatch("Installation done in").Report()
	}

	common.Log("Making micromamba executable ...")

	err := os.Chmod(BinMicromamba(), 0o755)
	if err != nil {
		common.Error("Install", err)
		return false
	}
	cloud.BackgroundMetric(common.ControllerIdentity(), "rcc.micromamba.install", common.Version)
	return true
}
