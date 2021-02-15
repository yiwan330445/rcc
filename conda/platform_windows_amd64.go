package conda

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"golang.org/x/sys/windows/registry"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/pretty"
	"github.com/robocorp/rcc/shell"
)

const (
	mingwSuffix             = "\\mingw-w64"
	Newline                 = "\r\n"
	defaultRobocorpLocation = "%LOCALAPPDATA%\\robocorp"
	librarySuffix           = "\\Library"
	scriptSuffix            = "\\Scripts"
	usrSuffix               = "\\bin"
	binSuffix               = "\\bin"
)

func MicromambaLink() string {
	return "https://downloads.robocorp.com/micromamba/v0.7.14/windows64/micromamba.exe"
}

var (
	Shell           = []string{"cmd.exe", "/K"}
	variablePattern = regexp.MustCompile("%[a-zA-Z]+%")
	FileExtensions  = []string{".exe", ".com", ".bat", ".cmd", ""}
)

func fromEnvironment(form string) string {
	replacement, ok := os.LookupEnv(form[1 : len(form)-1])
	if ok {
		return replacement
	}
	replacement, ok = os.LookupEnv(form)
	if ok {
		return replacement
	}
	return form
}

func ExpandPath(entry string) string {
	intermediate := os.ExpandEnv(entry)
	intermediate = variablePattern.ReplaceAllStringFunc(intermediate, fromEnvironment)
	result, err := filepath.Abs(intermediate)
	if err != nil {
		return intermediate
	}
	return result
}

func ensureHardlinkEnvironmment() (string, error) {
	return "", fmt.Errorf("Not implemented yet!")
}

func CondaEnvironment() []string {
	env := os.Environ()
	env = append(env, fmt.Sprintf("MAMBA_ROOT_PREFIX=%s", RobocorpHome()))
	tempFolder := RobocorpTemp()
	env = append(env, fmt.Sprintf("TEMP=%s", tempFolder))
	env = append(env, fmt.Sprintf("TMP=%s", tempFolder))
	return env
}

func BinMicromamba() string {
	return ExpandPath(filepath.Join(BinLocation(), "micromamba.exe"))
}

func CondaPaths(prefix string) []string {
	return []string{
		prefix,
		prefix + librarySuffix + mingwSuffix + binSuffix,
		prefix + librarySuffix + usrSuffix + binSuffix,
		prefix + librarySuffix + binSuffix,
		prefix + scriptSuffix,
		prefix + binSuffix,
	}
}

func IsPosix() bool {
	return false
}

func IsWindows() bool {
	return true
}

func HasLongPathSupport() bool {
	baseline := []string{RobocorpHome(), "stump"}
	stumpath := filepath.Join(baseline...)
	defer os.RemoveAll(stumpath)

	for count := 0; count < 24; count++ {
		baseline = append(baseline, fmt.Sprintf("verylongpath%d", count+1))
	}
	fullpath := filepath.Join(baseline...)

	code, err := shell.New(nil, ".", "cmd.exe", "/c", "mkdir", fullpath).Transparent()
	common.Trace("Checking long path support with MKDIR '%v' (%d characters) -> %v [%v] {%d}", fullpath, len(fullpath), err == nil, err, code)
	if err != nil {
		common.Log("%sWARNING!  Long path support failed. Reason: %v.%s", pretty.Red, err, pretty.Reset)
		common.Log("%sWARNING!  See %v for more details.%s", pretty.Red, longPathSupportArticle, pretty.Reset)
		return false
	}
	return true
}

func EnforceLongpathSupport() error {
	key, _, err := registry.CreateKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\FileSystem`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()
	return key.SetDWordValue("LongPathsEnabled", 1)
}
