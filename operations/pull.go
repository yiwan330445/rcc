package operations

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/robocorp/rcc/cloud"
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/fail"
	"github.com/robocorp/rcc/htfs"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/pretty"
	"github.com/robocorp/rcc/settings"
	"github.com/robocorp/rcc/xviper"
)

func pullOriginFingerprints(origin, catalogName string) (fingerprints string, count int, err error) {
	defer fail.Around(&err)

	client, err := cloud.NewClient(origin)
	fail.On(err != nil, "Could not create web client for %q, reason: %v", origin, err)

	request := client.NewRequest(fmt.Sprintf("/parts/%s", catalogName))
	response := client.Get(request)
	pretty.Guard(response.Status == 200, 5, "Problem with parts request, status=%d, body=%s", response.Status, response.Body)

	stream := bufio.NewReader(bytes.NewReader(response.Body))
	collection := make([]string, 0, 2048)
	for {
		line, err := stream.ReadString('\n')
		flat := strings.TrimSpace(line)
		if len(flat) > 0 {
			fullpath := htfs.ExactDefaultLocation(flat)
			if !pathlib.IsFile(fullpath) {
				collection = append(collection, flat)
			}
		}
		if err == io.EOF {
			return strings.Join(collection, "\n"), len(collection), nil
		}
		fail.On(err != nil, "STREAM error: %v", err)
	}

	return "", 0, fmt.Errorf("Unexpected reach of code that should never happen.")
}

func downloadMissingEnvironmentParts(origin, catalogName, selection string) (filename string, err error) {
	defer fail.Around(&err)

	url := fmt.Sprintf("%s/delta/%s", origin, catalogName)

	body := strings.NewReader(selection)
	filename = filepath.Join(os.TempDir(), fmt.Sprintf("peercc_%x.zip", os.Getpid()))

	client := &http.Client{Transport: settings.Global.ConfiguredHttpTransport()}
	request, err := http.NewRequest("POST", url, body)
	fail.On(err != nil, "Failed create request to %q failed, reason: %v", url, err)

	request.Header.Add("robocorp-installation-id", xviper.TrackingIdentity())
	request.Header.Add("User-Agent", common.UserAgent())

	response, err := client.Do(request)
	fail.On(err != nil, "Web request to %q failed, reason: %v", url, err)
	defer response.Body.Close()

	fail.On(response.StatusCode < 200 || 299 < response.StatusCode, "%s (%s)", response.Status, url)

	out, err := os.Create(filename)
	fail.On(err != nil, "Creating temporary file %q failed, reason: %v", filename, err)
	defer out.Close()
	defer pathlib.TryRemove("temporary", filename)

	digest := sha256.New()
	many := io.MultiWriter(out, digest)

	common.Debug("Downloading %s <%s> -> %s", url, response.Status, filename)

	_, err = io.Copy(many, response.Body)
	fail.On(err != nil, "Download failed, reason: %v", err)

	sum := fmt.Sprintf("%02x", digest.Sum(nil))
	finalname := filepath.Join(os.TempDir(), fmt.Sprintf("peercc_%s.zip", sum))
	err = pathlib.TryRename("delta", filename, finalname)
	fail.On(err != nil, "Rename %q -> %q failed, reason: %v", filename, finalname, err)

	return finalname, nil
}

func ProtectedImport(filename string) (err error) {
	defer fail.Around(&err)

	lockfile := common.HolotreeLock()
	completed := pathlib.LockWaitMessage(lockfile, "Serialized environment import [holotree lock]")
	locker, err := pathlib.Locker(lockfile, 30000)
	completed()
	fail.On(err != nil, "Could not get lock for holotree. Quiting.")
	defer locker.Release()

	common.Timeline("Import %v", filename)
	return Unzip(common.HololibLocation(), filename, true, false)
}

func PullCatalog(origin, catalogName string) (err error) {
	defer fail.Around(&err)

	common.Timeline("pull %q parts from %q", catalogName, origin)
	unknownSelected, count, err := pullOriginFingerprints(origin, catalogName)
	fail.On(err != nil, "%v", err)

	common.Timeline("download %d parts + catalog from %q", count, origin)
	filename, err := downloadMissingEnvironmentParts(origin, catalogName, unknownSelected)
	fail.On(err != nil, "%v", err)

	common.Debug("Temporary content based filename is: %q", filename)
	defer pathlib.TryRemove("temporary", filename)

	err = ProtectedImport(filename)
	fail.On(err != nil, "Failed to unzip %v to hololib, reason: %v", filename, err)
	common.Timeline("environment pull completed")

	return nil
}
