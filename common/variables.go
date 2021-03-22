package common

import (
	"fmt"
	"strings"
	"time"
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
	SemanticTag     string
	When            int64
	Clock           *stopwatch
	Settings        SettingsHold
)

type SettingsHold interface {
	DefaultEndpoint() string
	TelemetryURL() string
	IssuesURL() string
	PypiURL() string
	CondaURL() string
	DownloadsURL() string
}

func init() {
	Clock = &stopwatch{"Clock", time.Now()}
	When = Clock.started.Unix()
}

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
