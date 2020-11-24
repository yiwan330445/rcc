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
	Shell          = []string{"bash", "--noprofile", "--norc", "-i"}
	FileExtensions = []string{"", ".sh"}
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
	return []string{prefix + binSuffix}
}

func DownloadLink() string {
	return "https://repo.anaconda.com/miniconda/Miniconda3-latest-MacOSX-x86_64.sh"
}

func DownloadTarget() string {
	return filepath.Join(os.TempDir(), "miniconda3.sh")
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

func HasLongPathSupport() bool {
	return true
}

func ValidateLocations() bool {
	return true
}
