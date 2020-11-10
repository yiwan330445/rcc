package conda

import (
	"os"
	"path/filepath"
)

func DownloadLink() string {
	return "https://repo.anaconda.com/miniconda/Miniconda3-latest-Linux-x86.sh"
}

func DownloadTarget() string {
	return filepath.Join(os.TempDir(), "miniconda3.sh")
}
