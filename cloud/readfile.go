package cloud

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/pathlib"
)

func ReadFile(resource string) ([]byte, error) {
	link, err := url.ParseRequestURI(resource)
	if err != nil {
		return os.ReadFile(resource)
	}
	if link.Scheme == "file" || link.Scheme == "" {
		return os.ReadFile(link.Path)
	}
	tempfile := filepath.Join(pathlib.TempDir(), fmt.Sprintf("temp%x.part", common.When))
	defer os.Remove(tempfile)
	err = Download(resource, tempfile)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(tempfile)
}
