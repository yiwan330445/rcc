package conda

import (
	"os"
	"path/filepath"
)

const (
	mingwSuffix = "\\mingw-w64"
)

func DownloadLink() string {
	return "https://repo.anaconda.com/miniconda/Miniconda3-latest-Windows-x86_64.exe"
}

func DownloadTarget() string {
	return filepath.Join(os.TempDir(), "miniconda3.exe")
}
