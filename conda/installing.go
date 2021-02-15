package conda

import (
	"os"
	"time"

	"github.com/robocorp/rcc/cloud"
	"github.com/robocorp/rcc/common"
)

func MustMicromamba() bool {
	return HasMicroMamba() || ((DoDownload(1*time.Millisecond) || DoDownload(1*time.Second) || DoDownload(3*time.Second)) && DoInstall())
}

func DoDownload(delay time.Duration) bool {
	if common.DebugFlag {
		defer common.Stopwatch("Download done in").Report()
	}

	common.Log("Downloading micromamba, this may take awhile ...")

	time.Sleep(delay)

	err := DownloadMicromamba()
	if err != nil {
		common.Fatal("Download", err)
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
		common.Fatal("Install", err)
		return false
	}
	cloud.BackgroundMetric(common.ControllerIdentity(), "rcc.micromamba.install", common.Version)
	return true
}
