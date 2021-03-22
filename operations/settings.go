package operations

import (
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
	Settings       gateway
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

func CriticalEnvironmentSettingsCheck() {
	config, err := SummonSettings()
	pretty.Guard(err == nil, 80, "Aborting! Could not even get setting, reason: %v", err)
	result := &common.DiagnosticStatus{
		Details: make(map[string]string),
		Checks:  []*common.DiagnosticCheck{},
	}
	config.CriticalEnvironmentDiagnostics(result)
	diagnose := result.Diagnose("Settings")
	if HasCustomSettings() {
		diagnose.Ok("Uses custom settings at %q.", SettingsFileLocation())
	} else {
		diagnose.Ok("Uses builtin settings.")
	}
	fatal, fail, _, _ := result.Counts()
	if (fatal + fail) > 0 {
		humaneDiagnostics(os.Stderr, result)
		pretty.Guard(false, 111, "\nBroken settings.yaml. Cannot continue!")
	}
}

type gateway bool

func (it gateway) Endpoints() *settings.Endpoints {
	config, err := SummonSettings()
	pretty.Guard(err == nil, 111, "Could not get settings, reason: %v", err)
	pretty.Guard(config.Endpoints != nil, 111, "settings.yaml: endpoints are missing")
	return config.Endpoints
}

func (it gateway) DefaultEndpoint() string {
	return it.Endpoints().CloudApi
}

func (it gateway) IssuesURL() string {
	return it.Endpoints().Issues
}

func (it gateway) TelemetryURL() string {
	return it.Endpoints().Telemetry
}

func (it gateway) PypiURL() string {
	return it.Endpoints().Pypi
}

func (it gateway) CondaURL() string {
	return it.Endpoints().Conda
}

func (it gateway) DownloadsURL() string {
	return it.Endpoints().Downloads
}

func init() {
	Settings = gateway(true)
	common.Settings = Settings
}
