package common

import (
	"fmt"
	"strings"
)

var (
	Silent         bool
	DebugFlag      bool
	TraceFlag      bool
	NoCache        bool
	Liveonly       bool
	ControllerType string
)

const (
	DefaultEndpoint = "https://api.eu1.robocloud.eu/"
)

func UnifyVerbosityFlags() {
	if Silent {
		DebugFlag = false
		TraceFlag = false
	}
	if TraceFlag {
		DebugFlag = true
	}
}

func ForceDebug() {
	Silent = false
	DebugFlag = true
	UnifyVerbosityFlags()
}

func ControllerIdentity() string {
	return strings.ToLower(fmt.Sprintf("rcc.%s", ControllerType))
}
