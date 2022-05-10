package common

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/dchest/siphash"
)

const (
	FIXED_HOLOTREE                        = `FIXED_HOLOTREE`
	ROBOCORP_HOME_VARIABLE                = `ROBOCORP_HOME`
	VERBOSE_ENVIRONMENT_BUILDING          = `RCC_VERBOSE_ENVIRONMENT_BUILDING`
	ROBOCORP_OVERRIDE_SYSTEM_REQUIREMENTS = `ROBOCORP_OVERRIDE_SYSTEM_REQUIREMENTS`
)

var (
	Silent             bool
	DebugFlag          bool
	TraceFlag          bool
	DeveloperFlag      bool
	StrictFlag         bool
	LogLinenumbers     bool
	NoCache            bool
	NoOutputCapture    bool
	Liveonly           bool
	StageFolder        string
	ControllerType     string
	HolotreeSpace      string
	EnvironmentHash    string
	SemanticTag        string
	ForcedRobocorpHome string
	When               int64
	ProgressMark       time.Time
	Clock              *stopwatch
	randomIdentifier   string
)

func init() {
	Clock = &stopwatch{"Clock", time.Now()}
	When = Clock.started.Unix()
	ProgressMark = time.Now()

	randomIdentifier = fmt.Sprintf("%016x", rand.Uint64()^uint64(os.Getpid()))

	// Note: HololibCatalogLocation and HololibLibraryLocation are force
	//       created from "htfs" direcotry.go init function
	// Also: HolotreeLocation creation is left for actual holotree commands
	//       to prevent accidental access right problem during usage

	ensureDirectory(TemplateLocation())
	ensureDirectory(BinLocation())
	ensureDirectory(PipCache())
	ensureDirectory(WheelCache())
	ensureDirectory(RobotCache())
	ensureDirectory(MambaPackages())
}

func RobocorpHome() string {
	if len(ForcedRobocorpHome) > 0 {
		return ExpandPath(ForcedRobocorpHome)
	}
	home := os.Getenv(ROBOCORP_HOME_VARIABLE)
	if len(home) > 0 {
		return ExpandPath(home)
	}
	return ExpandPath(defaultRobocorpLocation)
}

func RobocorpLock() string {
	return filepath.Join(RobocorpHome(), "robocorp.lck")
}

func VerboseEnvironmentBuilding() bool {
	return DebugFlag || TraceFlag || len(os.Getenv(VERBOSE_ENVIRONMENT_BUILDING)) > 0
}

func OverrideSystemRequirements() bool {
	return len(os.Getenv(ROBOCORP_OVERRIDE_SYSTEM_REQUIREMENTS)) > 0
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
	return filepath.Join(RobocorpHome(), "templates")
}

func RobocorpTempRoot() string {
	return filepath.Join(RobocorpHome(), "temp")
}

func RobocorpTemp() string {
	tempLocation := filepath.Join(RobocorpTempRoot(), randomIdentifier)
	fullpath, err := filepath.Abs(tempLocation)
	if err != nil {
		fullpath = tempLocation
	}
	ensureDirectory(fullpath)
	if err != nil {
		Log("WARNING (%v) -> %v", tempLocation, err)
	}
	return fullpath
}

func BinLocation() string {
	return filepath.Join(RobocorpHome(), "bin")
}

func HoloLocation() string {
	return ExpandPath(defaultHoloLocation)
}

func FixedHolotreeLocation() bool {
	return len(os.Getenv(FIXED_HOLOTREE)) > 0
}

func HolotreeLocation() string {
	if FixedHolotreeLocation() {
		return HoloLocation()
	}
	return filepath.Join(RobocorpHome(), "holotree")
}

func HololibLocation() string {
	if FixedHolotreeLocation() {
		return filepath.Join(HoloLocation(), "lib")
	}
	return filepath.Join(RobocorpHome(), "hololib")
}

func HololibCatalogLocation() string {
	return filepath.Join(HololibLocation(), "catalog")
}

func HololibLibraryLocation() string {
	return filepath.Join(HololibLocation(), "library")
}

func HolotreeLock() string {
	return fmt.Sprintf("%s.lck", HolotreeLocation())
}

func UsesHolotree() bool {
	return len(HolotreeSpace) > 0
}

func PipCache() string {
	return filepath.Join(RobocorpHome(), "pipcache")
}

func WheelCache() string {
	return filepath.Join(RobocorpHome(), "wheels")
}

func RobotCache() string {
	return filepath.Join(RobocorpHome(), "robots")
}

func MambaRootPrefix() string {
	return RobocorpHome()
}

func MambaPackages() string {
	return ExpandPath(filepath.Join(MambaRootPrefix(), "pkgs"))
}

func PipRcFile() string {
	return ExpandPath(filepath.Join(RobocorpHome(), "piprc"))
}

func MicroMambaRcFile() string {
	return ExpandPath(filepath.Join(RobocorpHome(), "micromambarc"))
}

func SettingsFile() string {
	return ExpandPath(filepath.Join(RobocorpHome(), "settings.yaml"))
}

func CaBundleFile() string {
	return ExpandPath(filepath.Join(RobocorpHome(), "ca-bundle.pem"))
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

func Platform() string {
	return strings.ToLower(fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH))
}

func UserAgent() string {
	return fmt.Sprintf("rcc/%s (%s %s) %s", Version, runtime.GOOS, runtime.GOARCH, ControllerIdentity())
}

func ControllerIdentity() string {
	return strings.ToLower(fmt.Sprintf("rcc.%s", ControllerType))
}

func isDir(pathname string) bool {
	stat, err := os.Stat(pathname)
	return err == nil && stat.IsDir()
}

func ensureDirectory(name string) {
	if !isDir(name) {
		Error("mkdir", os.MkdirAll(name, 0o750))
	}
}

func UserHomeIdentity() string {
	location, err := os.UserHomeDir()
	if err != nil {
		return "badcafe"
	}
	digest := fmt.Sprintf("%02x", siphash.Hash(9007799254740993, 2147487647, []byte(location)))
	return digest[:7]
}
