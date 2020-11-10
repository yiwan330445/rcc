package conda

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/robocorp/rcc/common"
)

func DownloadConda() error {
	url := DownloadLink()
	filename := DownloadTarget()
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	digest := sha256.New()
	many := io.MultiWriter(out, digest)

	if common.Debug {
		common.Log("Downloading %s <%s> -> %s", url, response.Status, filename)
	}

	_, err = io.Copy(many, response.Body)
	if err != nil {
		return err
	}

	if common.Debug {
		sum := fmt.Sprintf("%02x", digest.Sum(nil))
		common.Log("SHA256 sum: %s", sum)
	}

	return nil
}
