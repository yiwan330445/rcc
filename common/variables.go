package common

import (
	"fmt"
	"strings"
)

var (
	Silent          bool
	DebugFlag       bool
	TraceFlag       bool
	NoCache         bool
	Liveonly        bool
	Stageonly       bool
	LeaseEffective  bool
	StageFolder     string
	ControllerType  string
	LeaseContract   string
	EnvironmentHash string
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

func UnifyStageHandling() {
	if len(StageFolder) > 0 {
		Liveonly = true
		Stageonly = true
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

func IsLeaseRequest() bool {
	return len(LeaseContract) > 0
}
