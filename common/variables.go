package common

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/robocorp/rcc/set"
)

type (
	Verbosity uint8
)

const (
	Undefined Verbosity = 0
	Silently  Verbosity = 1
	Normal    Verbosity = 2
	Debugging Verbosity = 3
	Tracing   Verbosity = 4
)

const (
	ROBOCORP_HOME_VARIABLE                = `ROBOCORP_HOME`
	RCC_REMOTE_ORIGIN                     = `RCC_REMOTE_ORIGIN`
	RCC_REMOTE_AUTHORIZATION              = `RCC_REMOTE_AUTHORIZATION`
	RCC_NO_TEMP_MANAGEMENT                = `RCC_NO_TEMP_MANAGEMENT`
	RCC_NO_PYC_MANAGEMENT                 = `RCC_NO_PYC_MANAGEMENT`
	VERBOSE_ENVIRONMENT_BUILDING          = `RCC_VERBOSE_ENVIRONMENT_BUILDING`
	ROBOCORP_OVERRIDE_SYSTEM_REQUIREMENTS = `ROBOCORP_OVERRIDE_SYSTEM_REQUIREMENTS`
	RCC_VERBOSITY                         = `RCC_VERBOSITY`
	SILENTLY                              = `silent`
	TRACING                               = `trace`
	DEBUGGING                             = `debug`
)

var (
	NoBuild                 bool
	NoTempManagement        bool
	NoPycManagement         bool
	ExternallyManaged       bool
	DeveloperFlag           bool
	StrictFlag              bool
	SharedHolotree          bool
	LogLinenumbers          bool
	NoCache                 bool
	NoOutputCapture         bool
	Liveonly                bool
	UnmanagedSpace          bool
	FreshlyBuildEnvironment bool
	WarrantyVoidedFlag      bool
	StageFolder             string
	ControllerType          string
	HolotreeSpace           string
	EnvironmentHash         string
	SemanticTag             string
	ForcedRobocorpHome      string
	When                    int64
	Clock                   *stopwatch
	randomIdentifier        string
	verbosity               Verbosity
	LogHides                []string
)

func init() {
	Clock = &stopwatch{"Clock", time.Now()}
	When = Clock.When()

	randomIdentifier = fmt.Sprintf("%016x", rand.Uint64()^uint64(os.Getpid()))

	lowargs := make([]string, 0, len(os.Args))
	for _, arg := range os.Args {
		lowargs = append(lowargs, strings.ToLower(arg))
	}
	// peek CLI options to pre-initialize "Warranty Voided" indicator
	args := set.Set(lowargs)
	WarrantyVoidedFlag = set.Member(args, "--warranty-voided")
	if set.Member(args, "--debug") {
		verbosity = Debugging
	}
	if set.Member(args, "--trace") {
		verbosity = Tracing
	}
	if set.Member(args, "--no-temp-management") {
		NoTempManagement = true
	}
	if set.Member(args, "--no-pyc-management") {
		NoPycManagement = true
	}

	// Note: HololibCatalogLocation, HololibLibraryLocation and HololibUsageLocation
	//       are force created from "htfs" direcotry.go init function
	// Also: HolotreeLocation creation is left for actual holotree commands
	//       to prevent accidental access right problem during usage

	SharedHolotree = isFile(HoloInitUserFile())

	ensureDirectory(JournalLocation())
	ensureDirectory(TemplateLocation())
	ensureDirectory(BinLocation())
	ensureDirectory(PipCache())
	ensureDirectory(WheelCache())
	ensureDirectory(RobotCache())
	ensureDirectory(MambaPackages())
}

func RandomIdentifier() string {
	return randomIdentifier
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

func DisableTempManagement() bool {
	return NoTempManagement || len(os.Getenv(RCC_NO_TEMP_MANAGEMENT)) > 0
}

func DisablePycManagement() bool {
	return NoPycManagement || len(os.Getenv(RCC_NO_PYC_MANAGEMENT)) > 0
}

func RccRemoteOrigin() string {
	return os.Getenv(RCC_REMOTE_ORIGIN)
}

func RccRemoteAuthorization() (string, bool) {
	result := os.Getenv(RCC_REMOTE_AUTHORIZATION)
	return result, len(result) > 0
}

func RobocorpLock() string {
	return filepath.Join(RobocorpHome(), "robocorp.lck")
}

func WarrantyVoided() bool {
	return WarrantyVoidedFlag
}

func DebugFlag() bool {
	return verbosity >= Debugging
}

func TraceFlag() bool {
	return verbosity >= Tracing
}

func VerboseEnvironmentBuilding() bool {
	return DebugFlag() || TraceFlag() || len(os.Getenv(VERBOSE_ENVIRONMENT_BUILDING)) > 0
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

func OldEventJournal() string {
	return filepath.Join(RobocorpHome(), "event.log")
}

func EventJournal() string {
	return filepath.Join(JournalLocation(), "event.log")
}

func JournalLocation() string {
	return filepath.Join(RobocorpHome(), "journals")
}

func TemplateLocation() string {
	return filepath.Join(RobocorpHome(), "templates")
}

func RobocorpTempRoot() string {
	return filepath.Join(RobocorpHome(), "temp")
}

func RobocorpTempName() string {
	return filepath.Join(RobocorpTempRoot(), RandomIdentifier())
}

func RobocorpTemp() string {
	tempLocation := RobocorpTempName()
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

func MicromambaLocation() string {
	return filepath.Join(RobocorpHome(), "micromamba")
}

func SharedMarkerLocation() string {
	return filepath.Join(HoloLocation(), "shared.yes")
}

func HoloLocation() string {
	return ExpandPath(defaultHoloLocation)
}

func HoloInitLocation() string {
	return filepath.Join(HoloLocation(), "lib", "catalog", "init")
}

func HoloInitUserFile() string {
	return filepath.Join(HoloInitLocation(), UserHomeIdentity())
}

func HoloInitCommonFile() string {
	return filepath.Join(HoloInitLocation(), "commons.tof")
}

func HolotreeLocation() string {
	if SharedHolotree {
		return HoloLocation()
	}
	return filepath.Join(RobocorpHome(), "holotree")
}

func HololibLocation() string {
	if SharedHolotree {
		return filepath.Join(HoloLocation(), "lib")
	}
	return filepath.Join(RobocorpHome(), "hololib")
}

func HololibPids() string {
	return filepath.Join(HololibLocation(), "pids")
}

func HololibCatalogLocation() string {
	return filepath.Join(HololibLocation(), "catalog")
}

func HololibLibraryLocation() string {
	return filepath.Join(HololibLocation(), "library")
}

func HololibUsageLocation() string {
	return filepath.Join(HololibLocation(), "used")
}

func HolotreeLock() string {
	return filepath.Join(HolotreeLocation(), "global.lck")
}

func BadHololibSitePackagesLocation() string {
	return filepath.Join(HololibLocation(), "site-packages")
}

func BadHololibScriptsLocation() string {
	if SharedHolotree {
		return filepath.Join(HoloLocation(), "Scripts")
	}
	return filepath.Join(RobocorpHome(), "Scripts")
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

func CaBundleDir() string {
	return ExpandPath(RobocorpHome())
}

func DefineVerbosity(silent, debug, trace bool) {
	override := os.Getenv(RCC_VERBOSITY)
	switch {
	case silent || override == SILENTLY:
		verbosity = Silently
	case trace || override == TRACING:
		verbosity = Tracing
	case debug || override == DEBUGGING:
		verbosity = Debugging
	default:
		verbosity = Normal
	}
}

func Silent() bool {
	return verbosity == Silently
}

func UnifyStageHandling() {
	if len(StageFolder) > 0 {
		Liveonly = true
	}
}

func ForceDebug() {
	DefineVerbosity(false, true, false)
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

func isFile(pathname string) bool {
	stat, err := os.Stat(pathname)
	return err == nil && stat.Mode().IsRegular()
}

func isDir(pathname string) bool {
	stat, err := os.Stat(pathname)
	return err == nil && stat.IsDir()
}

func ensureDirectory(name string) {
	if !WarrantyVoided() && !isDir(name) {
		Error("mkdir", os.MkdirAll(name, 0o750))
	}
}

func SymbolicUserIdentity() string {
	location, err := os.UserHomeDir()
	if err != nil {
		return "badcafe"
	}
	digest := fmt.Sprintf("%02x", Siphash(9007799254740993, 2147487647, []byte(location)))
	return digest[:7]
}

func UserHomeIdentity() string {
	if UnmanagedSpace {
		return "UNMNGED"
	}
	return SymbolicUserIdentity()
}
