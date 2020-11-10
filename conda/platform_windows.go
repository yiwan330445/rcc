package conda

import (
	"os"
	"path/filepath"
	"regexp"
)

const (
	Newline                 = "\r\n"
	defaultRobocorpLocation = "%LOCALAPPDATA%\\robocorp"
	librarySuffix           = "\\Library"
	scriptSuffix            = "\\Scripts"
	usrSuffix               = "\\bin"
	binSuffix               = "\\bin"
)

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

func BinConda() string {
	return ExpandPath(filepath.Join(MinicondaLocation(), "Scripts", "conda.exe"))
}

func BinPython() string {
	return ExpandPath(filepath.Join(MinicondaLocation(), "python.exe"))
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

func InstallCommand() []string {
	return []string{DownloadTarget(), "/InstallationType=JustMe", "/NoRegisty=1", "/S", "/D=" + MinicondaLocation()}
}

func IsPosix() bool {
	return false
}

func IsWindows() bool {
	return true
}

func ValidateLocations() bool {
	checked := map[string]string{
		"Environment variable 'TMP'":           os.Getenv("TMP"),
		"Environment variable 'TEMP'":          os.Getenv("TEMP"),
		"Environment variable 'ROBOCORP_HOME'": os.Getenv("ROBOCORP_HOME"),
		"Path to 'ROBOCORP_HOME' directory":    RobocorpHome(),
	}
	return validateLocations(checked)
}
