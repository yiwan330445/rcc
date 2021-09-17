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
	StageFolder     string
	ControllerType  string
	HolotreeSpace   string
	EnvironmentHash string
	SemanticTag     string
	When            int64
	ProgressMark    time.Time
	Clock           *stopwatch
)

func init() {
	Clock = &stopwatch{"Clock", time.Now()}
	When = Clock.started.Unix()
	ProgressMark = time.Now()
}

func RobocorpHome() string {
	home := os.Getenv(ROBOCORP_HOME_VARIABLE)
	if len(home) > 0 {
		return ensureDirectory(ExpandPath(home))
	}
	return ensureDirectory(ExpandPath(defaultRobocorpLocation))
}

func RobocorpLock() string {
	return filepath.Join(RobocorpHome(), "robocorp.lck")
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

func BinRcc() string {
	self, err := os.Executable()
	if err != nil {
		return os.Args[0]
	}
	return self
}

func EventJournal() string {
	return filepath.Join(RobocorpHome(), "event.log")
}

func TemplateLocation() string {
	return ensureDirectory(filepath.Join(RobocorpHome(), "templates"))
}

func RobocorpTempRoot() string {
	return filepath.Join(RobocorpHome(), "temp")
}

func BinLocation() string {
	return ensureDirectory(filepath.Join(RobocorpHome(), "bin"))
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

func MambaPackages() string {
	return ExpandPath(filepath.Join(RobocorpHome(), "pkgs"))
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
