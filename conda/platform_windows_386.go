package conda

import (
	"os"
	"path/filepath"
)

const (
	mingwSuffix = "\\mingw-w32"
)

func DownloadLink() string {
	return "https://repo.anaconda.com/miniconda/Miniconda3-latest-Windows-x86.exe"
}

func DownloadTarget() string {
	return filepath.Join(os.TempDir(), "miniconda3.exe")
}
