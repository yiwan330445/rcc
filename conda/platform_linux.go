package conda

import (
	"os"
	"path/filepath"
)

const (
	Newline                 = "\n"
	defaultRobocorpLocation = "$HOME/.robocorp"
	binSuffix               = "/bin"
)

var (
	FileExtensions = []string{""}
	Shell          = []string{"bash", "--noprofile", "--norc", "-i"}
)

func ExpandPath(entry string) string {
	intermediate := os.ExpandEnv(entry)
	result, err := filepath.Abs(intermediate)
	if err != nil {
		return intermediate
	}
	return result
}

func BinConda() string {
	return ExpandPath(filepath.Join(MinicondaLocation(), "bin", "conda"))
}

func BinPython() string {
	return ExpandPath(filepath.Join(MinicondaLocation(), "bin", "python"))
}

func CondaPaths(prefix string) []string {
	return []string{ExpandPath(prefix + binSuffix)}
}

func InstallCommand() []string {
	return []string{"bash", DownloadTarget(), "-u", "-b", "-p", MinicondaLocation()}
}

func IsPosix() bool {
	return true
}

func IsWindows() bool {
	return false
}

func ValidateLocations() bool {
	return true
}
