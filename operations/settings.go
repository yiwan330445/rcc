package operations

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/robocorp/rcc/blobs"
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/pretty"
	"github.com/robocorp/rcc/settings"
)

var (
	cachedSettings *settings.Settings
)

func cacheSettings(result *settings.Settings) (*settings.Settings, error) {
	if result != nil {
		cachedSettings = result
	}
	return result, nil
}

func SettingsFileLocation() string {
	return filepath.Join(conda.RobocorpHome(), "settings.yaml")
}

func HasCustomSettings() bool {
	return pathlib.IsFile(SettingsFileLocation())
}

func DefaultSettings() ([]byte, error) {
	return blobs.Asset("assets/settings.yaml")
}

func rawSettings() (content []byte, location string, err error) {
	if HasCustomSettings() {
		location = SettingsFileLocation()
		content, err = ioutil.ReadFile(location)
		return content, location, err
	} else {
		content, err = DefaultSettings()
		return content, "builtin", err
	}
}

func SummonSettings() (*settings.Settings, error) {
	if cachedSettings != nil {
		return cachedSettings, nil
	}
	content, source, err := rawSettings()
	if err != nil {
		return nil, err
	}
	config, err := settings.FromBytes(content)
	if err != nil {
		return nil, err
	}
	return cacheSettings(config.Source(source))
}

func DefaultEndpoint() (string, error) {
	config, err := SummonSettings()
	if err != nil {
		return "", err
	}
	endpoints := config.Endpoints
	if endpoints == nil {
		return "", fmt.Errorf("Brokens settings: all endpoints are missing!")
	}
	return "", nil
}

func CriticalEnvironmentSettingsCheck() {
	return
	// JIPPO:FIXME:JIPPO -- continue here
	config, err := SummonSettings()
	pretty.Guard(err == nil, 80, "Aborting! Could not even get setting, reason: %v", err)
	result := &common.DiagnosticStatus{
		Details: make(map[string]string),
		Checks:  []*common.DiagnosticCheck{},
	}
	config.Diagnostics(result)
	humaneDiagnostics(os.Stderr, result)
}
