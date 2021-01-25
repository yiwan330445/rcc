package conda

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/shell"
)

const (
	ROBOCORP_HOME_VARIABLE = `ROBOCORP_HOME`
)

var (
	ignoredPaths     = []string{"python", "conda"}
	pythonPaths      = []string{"resources", "libraries", "tasks", "variables"}
	hashPattern      = regexp.MustCompile("^[0-9a-f]{16}(?:\\.meta)?$")
	randomIdentifier string
)

func init() {
	randomIdentifier = fmt.Sprintf("%016x", rand.Uint64()^uint64(os.Getpid()))
}

func sorted(files []os.FileInfo) {
	sort.SliceStable(files, func(left, right int) bool {
		return files[left].Name() < files[right].Name()
	})
}

func DigestFor(folder string) ([]byte, error) {
	handle, err := os.Open(folder)
	if err != nil {
		return nil, err
	}
	defer handle.Close()
	entries, err := handle.Readdir(-1)
	if err != nil {
		return nil, err
	}
	digester := sha256.New()
	sorted(entries)
	for _, entry := range entries {
		if entry.IsDir() {
			if entry.Name() == "__pycache__" {
				continue
			}
			digest, err := DigestFor(filepath.Join(folder, entry.Name()))
			if err != nil {
				return nil, err
			}
			digester.Write(digest)
			continue
		}
		repr := fmt.Sprintf("%s -- %x", entry.Name(), entry.Size())
		digester.Write([]byte(repr))
	}
	result := digester.Sum([]byte{})
	return result, nil
}

func hashedEntity(name string) bool {
	return hashPattern.MatchString(name)
}

func hasDatadir(basedir, metafile string) bool {
	if filepath.Ext(metafile) != ".meta" {
		return false
	}
	fullpath := filepath.Join(basedir, metafile)
	stat, err := os.Stat(fullpath[:len(fullpath)-5])
	return err == nil && stat.IsDir()
}

func hasMetafile(basedir, subdir string) bool {
	folder := filepath.Join(basedir, subdir)
	_, err := os.Stat(metafile(folder))
	return err == nil
}

func dirnamesFrom(location string) []string {
	result := make([]string, 0, 20)
	handle, err := os.Open(ExpandPath(location))
	if err != nil {
		common.Error("Warning", err)
		return result
	}
	defer handle.Close()
	children, err := handle.Readdir(-1)
	if err != nil {
		common.Error("Warning", err)
		return result
	}

	for _, child := range children {
		if child.IsDir() && hasMetafile(location, child.Name()) {
			result = append(result, child.Name())
		}
	}

	return result
}

func orphansFrom(location string) []string {
	result := make([]string, 0, 20)
	handle, err := os.Open(ExpandPath(location))
	if err != nil {
		common.Error("Warning", err)
		return result
	}
	defer handle.Close()
	children, err := handle.Readdir(-1)
	if err != nil {
		common.Error("Warning", err)
		return result
	}

	for _, child := range children {
		hashed := hashedEntity(child.Name())
		if hashed && child.IsDir() && hasMetafile(location, child.Name()) {
			continue
		}
		if hashed && !child.IsDir() && hasDatadir(location, child.Name()) {
			continue
		}
		result = append(result, filepath.Join(location, child.Name()))
	}

	return result
}

func FindPath(environment string) pathlib.PathParts {
	target := pathlib.TargetPath()
	target = target.Remove(ignoredPaths)
	target = target.Prepend(CondaPaths(environment)...)
	return target
}

func PythonPath() pathlib.PathParts {
	return pathlib.PathFrom(pythonPaths...)
}

func EnvironmentExtensionFor(location string) []string {
	environment := make([]string, 0, 20)
	searchPath := FindPath(location)
	python, ok := searchPath.Which("python3", FileExtensions)
	if !ok {
		python, ok = searchPath.Which("python", FileExtensions)
	}
	if ok {
		environment = append(environment, "PYTHON_EXE="+python)
	}
	return append(environment,
		"CONDA_DEFAULT_ENV=rcc",
		"CONDA_PREFIX="+location,
		"CONDA_PROMPT_MODIFIER=(rcc)",
		"CONDA_SHLVL=1",
		"PYTHONHOME=",
		"PYTHONSTARTUP=",
		"PYTHONEXECUTABLE=",
		"PYTHONNOUSERSITE=1",
		"ROBOCORP_HOME="+RobocorpHome(),
		"RCC_ENVIRONMENT_HASH="+common.EnvironmentHash,
		"TEMP="+RobocorpTemp(),
		"TMP="+RobocorpTemp(),
		searchPath.AsEnvironmental("PATH"),
		PythonPath().AsEnvironmental("PYTHONPATH"),
	)
}

func EnvironmentFor(location string) []string {
	return append(os.Environ(), EnvironmentExtensionFor(location)...)
}

func MambaPackages() string {
	return ExpandPath(filepath.Join(RobocorpHome(), "pkgs"))
}

func MambaCache() string {
	return ExpandPath(filepath.Join(MambaPackages(), "cache"))
}

func asVersion(text string) (uint64, string) {
	text = strings.TrimSpace(text)
	parts := strings.SplitN(text, ".", 4)
	steps := len(parts)
	multipliers := []uint64{1000000, 1000, 1}
	version := uint64(0)
	for at, multiplier := range multipliers {
		if steps <= at {
			break
		}
		value, err := strconv.ParseUint(parts[at], 10, 64)
		if err != nil {
			break
		}
		version += multiplier * value
	}
	return version, text
}

func MicromambaVersion() string {
	versionText, _, err := shell.New(CondaEnvironment(), ".", BinMicromamba(), "--version").CaptureOutput()
	if err != nil {
		return err.Error()
	}
	_, versionText = asVersion(versionText)
	return versionText
}

func HasMicroMamba() bool {
	if !pathlib.IsFile(BinMicromamba()) {
		return false
	}
	version, versionText := asVersion(MicromambaVersion())
	goodEnough := version >= 7010
	common.Debug("%q version is %q -> %v (good enough: %v)", BinMicromamba(), versionText, version, goodEnough)
	return goodEnough
}

func RobocorpHome() string {
	home := os.Getenv(ROBOCORP_HOME_VARIABLE)
	if len(home) > 0 {
		return ExpandPath(home)
	}
	return ExpandPath(defaultRobocorpLocation)
}

func RobocorpTempRoot() string {
	return filepath.Join(RobocorpHome(), "temp")
}

func RobocorpTemp() string {
	tempLocation := filepath.Join(RobocorpTempRoot(), randomIdentifier)
	fullpath, err := pathlib.EnsureDirectory(tempLocation)
	if err != nil {
		common.Log("WARNING (%v) -> %v", tempLocation, err)
	}
	return fullpath
}

func BinLocation() string {
	return filepath.Join(RobocorpHome(), "bin")
}

func LiveLocation() string {
	return ensureDirectory(filepath.Join(RobocorpHome(), "live"))
}

func TemplateLocation() string {
	return ensureDirectory(filepath.Join(RobocorpHome(), "base"))
}

func RobocorpLock() string {
	return fmt.Sprintf("%s.lck", LiveLocation())
}

func MinicondaLocation() string {
	// Legacy function, but must remain until cleanup is done
	return filepath.Join(RobocorpHome(), "miniconda3")
}

func ensureDirectory(name string) string {
	pathlib.EnsureDirectoryExists(name)
	return name
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

func LocalChannel() (string, bool) {
	basefolder := filepath.Join(RobocorpHome(), "channel")
	fullpath := filepath.Join(basefolder, "channeldata.json")
	stats, err := os.Stat(fullpath)
	if err != nil {
		return "", false
	}
	if !stats.IsDir() {
		return basefolder, true
	}
	return "", false
}

func TemplateFrom(hash string) string {
	return filepath.Join(TemplateLocation(), hash)
}

func LiveFrom(hash string) string {
	if common.Stageonly {
		return common.StageFolder
	}
	return ExpandPath(filepath.Join(LiveLocation(), hash))
}

func TemplateList() []string {
	return dirnamesFrom(TemplateLocation())
}

func LiveList() []string {
	return dirnamesFrom(LiveLocation())
}

func OrphanList() []string {
	result := orphansFrom(TemplateLocation())
	result = append(result, orphansFrom(LiveLocation())...)
	return result
}
