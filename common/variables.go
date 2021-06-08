package common

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	ROBOCORP_HOME_VARIABLE                = `ROBOCORP_HOME`
	VERBOSE_ENVIRONMENT_BUILDING          = `RCC_VERBOSE_ENVIRONMENT_BUILDING`
	ROBOCORP_OVERRIDE_SYSTEM_REQUIREMENTS = `ROBOCORP_OVERRIDE_SYSTEM_REQUIREMENTS`
)

var (
	Silent          bool
	DebugFlag       bool
	TraceFlag       bool
	LogLinenumbers  bool
	NoCache         bool
	NoOutputCapture bool
	Liveonly        bool
	Stageonly       bool
	LeaseEffective  bool
	StageFolder     string
	ControllerType  string
	HolotreeSpace   string
	LeaseContract   string
	EnvironmentHash string
	SemanticTag     string
	When            int64
	Clock           *stopwatch
)

func init() {
	Clock = &stopwatch{"Clock", time.Now()}
	When = Clock.started.Unix()
}

func RobocorpHome() string {
	home := os.Getenv(ROBOCORP_HOME_VARIABLE)
	if len(home) > 0 {
		return ensureDirectory(ExpandPath(home))
	}
	return ensureDirectory(ExpandPath(defaultRobocorpLocation))
}

func RobocorpLock() string {
	return fmt.Sprintf("%s.lck", LiveLocation())
}

func VerboseEnvironmentBuilding() bool {
	return len(os.Getenv(VERBOSE_ENVIRONMENT_BUILDING)) > 0
}

func OverrideSystemRequirements() bool {
	return len(os.Getenv(ROBOCORP_OVERRIDE_SYSTEM_REQUIREMENTS)) > 0
}

func ensureDirectory(name string) string {
	os.MkdirAll(name, 0o750)
	return name
}

func EventJournal() string {
	return filepath.Join(RobocorpHome(), "event.log")
}

func TemplateLocation() string {
	return ensureDirectory(filepath.Join(RobocorpHome(), "templates"))
}

func BinLocation() string {
	return ensureDirectory(filepath.Join(RobocorpHome(), "bin"))
}

func LiveLocation() string {
	return ensureDirectory(filepath.Join(RobocorpHome(), "live"))
}

func BaseLocation() string {
	return ensureDirectory(filepath.Join(RobocorpHome(), "base"))
}

func HololibLocation() string {
	return ensureDirectory(filepath.Join(RobocorpHome(), "hololib"))
}

func HololibCatalogLocation() string {
	return ensureDirectory(filepath.Join(HololibLocation(), "catalog"))
}

func HololibLibraryLocation() string {
	return ensureDirectory(filepath.Join(HololibLocation(), "library"))
}

func HolotreeLock() string {
	return fmt.Sprintf("%s.lck", HolotreeLocation())
}

func HolotreeLocation() string {
	return ensureDirectory(filepath.Join(RobocorpHome(), "holotree"))
}

func UsesHolotree() bool {
	return len(HolotreeSpace) > 0
}

func PipCache() string {
	return ensureDirectory(filepath.Join(RobocorpHome(), "pipcache"))
}

func WheelCache() string {
	return ensureDirectory(filepath.Join(RobocorpHome(), "wheels"))
}

func RobotCache() string {
	return ensureDirectory(filepath.Join(RobocorpHome(), "robots"))
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

func UserAgent() string {
	return fmt.Sprintf("rcc/%s (%s %s) %s", Version, runtime.GOOS, runtime.GOARCH, ControllerIdentity())
}

func ControllerIdentity() string {
	return strings.ToLower(fmt.Sprintf("rcc.%s", ControllerType))
}

func IsLeaseRequest() bool {
	return len(LeaseContract) > 0
}
