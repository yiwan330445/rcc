package conda

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/settings"
)

func DownloadMicromamba() error {
	common.Timeline("downloading micromamba")
	url := MicromambaLink()
	filename := BinMicromamba()
	client := &http.Client{Transport: settings.Global.ConfiguredHttpTransport()}
	response, err := client.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("Downloading %q failed, reason: %q!", url, response.Status)
	}

	if pathlib.Exists(BinMicromamba()) {
		os.Remove(BinMicromamba())
	}

	pathlib.EnsureDirectory(filepath.Dir(BinMicromamba()))
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	digest := sha256.New()
	many := io.MultiWriter(out, digest)

	common.Debug("Downloading %s <%s> -> %s", url, response.Status, filename)

	_, err = io.Copy(many, response.Body)
	if err != nil {
		return err
	}

	return common.Debug("SHA256 sum: %02x", digest.Sum(nil))
}
