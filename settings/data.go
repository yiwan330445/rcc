package settings

import (
	"encoding/json"

	"github.com/robocorp/rcc/common"
	"gopkg.in/yaml.v1"
)

type StringMap map[string]string

type Settings struct {
	Autoupdates  StringMap     `yaml:"autoupdates" json:"autoupdates"`
	Branding     StringMap     `yaml:"branding" json:"branding"`
	BusinessData *BusinessData `yaml:"business-data" json:"business-data"`
	Certificates *Certificates `yaml:"certificates" json:"certificates"`
	Endpoints    *Endpoints    `yaml:"endpoints" json:"endpoints"`
	Logs         *Logs         `yaml:"logs" json:"logs"`
	Meta         *Meta         `yaml:"meta" json:"meta"`
	Proxies      *Proxies      `yaml:"proxies" json:"proxies"`
}

func FromBytes(raw []byte) (*Settings, error) {
	var settings Settings
	err := yaml.Unmarshal(raw, &settings)
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

func (it *Settings) AsYaml() ([]byte, error) {
	content, err := yaml.Marshal(it)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (it *Settings) Source(filename string) *Settings {
	if it.Meta != nil && len(filename) > 0 {
		it.Meta.Source = filename
	}
	return it
}

func (it *Settings) AsJson() ([]byte, error) {
	content, err := json.MarshalIndent(it, "", "  ")
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (it *Settings) Diagnostics(target *common.DiagnosticStatus) {
	diagnose := target.Diagnose("Settings")
	correct := true
	if it.BusinessData == nil {
		diagnose.Warning("", "settings.yaml: business-data section is totally missing")
		correct = false
	}
	if it.Certificates == nil {
		diagnose.Warning("", "settings.yaml: certificates section is totally missing")
		correct = false
	}
	if it.Endpoints == nil {
		correct = false
	}
	if it.Logs == nil {
		diagnose.Warning("", "settings.yaml: logs section is totally missing")
		correct = false
	}
	if it.Meta == nil {
		diagnose.Warning("", "settings.yaml: meta section is totally missing")
		correct = false
	}
	if it.Proxies == nil {
		diagnose.Warning("", "settings.yaml: proxies section is totally missing")
		correct = false
	}
	if correct {
		diagnose.Ok("Toplevel settings are ok.")
	}
}

type BusinessData struct {
	RootLocation string `yaml:"root-location" json:"root-location"`
}

type Certificates struct {
	IgnoreVerification bool   `yaml:"ignore-ssl-verification" json:"ignore-ssl-verification"`
	RootLocation       string `yaml:"root-location" json:"root-location"`
}

type Endpoints struct {
	CloudApi     string `yaml:"cloud-api" json:"cloud-api"`
	CloudLinking string `yaml:"cloud-linking" json:"cloud-linking"`
	Conda        string `yaml:"conda" json:"conda"`
	Docs         string `yaml:"docs" json:"docs"`
	Downloads    string `yaml:"downloads" json:"downloads"`
	Issues       string `yaml:"issues" json:"issues"`
	Portal       string `yaml:"portal" json:"portal"`
	Pypi         string `yaml:"pypi" json:"pypi"`
	RobotPull    string `yaml:"robot-pull" json:"robot-pull"`
	Telemetry    string `yaml:"telemetry" json:"telemetry"`
}

type Logs struct {
	Level        string `yaml:"level" json:"level"`
	RootLocation string `yaml:"root-location" json:"root-location"`
}

type Meta struct {
	Source  string `yaml:"source" json:"source"`
	Version string `yaml:"version" json:"version"`
}

type Proxies struct {
	Http  string `yaml:"http" json:"http"`
	Https string `yaml:"https" json:"https"`
}
