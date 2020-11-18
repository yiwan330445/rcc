package pretty

import (
	"os"

	"github.com/mattn/go-isatty"
	"github.com/robocorp/rcc/common"
)

var (
	Disabled    bool
	Interactive bool
	White       string
	Grey        string
	Black       string
	Red         string
	Green       string
	Yellow      string
	Cyan        string
	Reset       string
)

func Setup() {
	stdin := isatty.IsTerminal(os.Stdin.Fd())
	stdout := isatty.IsTerminal(os.Stdout.Fd())
	stderr := isatty.IsTerminal(os.Stderr.Fd())
	Interactive = stdin && stdout && stderr

	localSetup()

	common.Trace("Interactive mode enabled: %v; colors enabled: %v", Interactive, !Disabled)
	if Interactive && !Disabled {
		White = csi("97m")
		Grey = csi("90m")
		Black = csi("30m")
		Red = csi("91m")
		Green = csi("92m")
		Cyan = csi("96m")
		Yellow = csi("93m")
		Reset = csi("0m")
	}
}
