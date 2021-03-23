package settings

import (
	"encoding/json"
	"net/url"
	"sort"
	"strings"

	"github.com/robocorp/rcc/common"
	"gopkg.in/yaml.v1"
)

const (
	httpsPrefix = `https://`
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

func diagnoseUrl(link, label string, diagnose common.Diagnoser, correct bool) bool {
	if len(link) == 0 {
		diagnose.Fatal("", "required %q URL is missing.", label)
		return false
	}
	if !strings.HasPrefix(link, httpsPrefix) {
		diagnose.Fatal("", "%q URL %q is does not start with %q prefix.", label, link, httpsPrefix)
		return false
	}
	_, err := url.Parse(link)
	if err != nil {
		diagnose.Fatal("", "%q URL %q cannot be parsed, reason %v.", label, link, err)
		return false
	}
	return correct
}

func (it *Settings) CriticalEnvironmentDiagnostics(target *common.DiagnosticStatus) {
	diagnose := target.Diagnose("settings.yaml")
	correct := true
	if it.Endpoints == nil {
		diagnose.Fatal("", "endpoints section is totally missing")
		correct = false
	} else {
		correct = diagnoseUrl(it.Endpoints.CloudApi, "endpoints/cloud-api", diagnose, correct)
		correct = diagnoseUrl(it.Endpoints.Conda, "endpoints/conda", diagnose, correct)
		correct = diagnoseUrl(it.Endpoints.Pypi, "endpoints/pypi", diagnose, correct)
		correct = diagnoseUrl(it.Endpoints.Downloads, "endpoints/downloads", diagnose, correct)
	}
	if correct {
		diagnose.Ok("Toplevel settings are ok.")
	}
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
		diagnose.Warning("", "settings.yaml: endpoints section is totally missing")
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
	VerifySsl    bool   `yaml:"verify-ssl" json:"verify-ssl"`
	RootLocation string `yaml:"root-location" json:"root-location"`
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
	PypiFiles    string `yaml:"pypi-files" json:"pypi-files"`
	PypiTrusted  string `yaml:"pypi-trusted" json:"pypi-trusted"`
	RobotPull    string `yaml:"robot-pull" json:"robot-pull"`
	Telemetry    string `yaml:"telemetry" json:"telemetry"`
}

func hostFromUrl(link string, collector map[string]bool) {
	if len(link) == 0 {
		return
	}
	parsed, err := url.Parse(link)
	if err != nil {
		return
	}
	parts := strings.SplitN(parsed.Host, ":", 2)
	collector[parts[0]] = true
}

func (it *Endpoints) Hosts() []string {
	collector := make(map[string]bool)
	hostFromUrl(it.CloudApi, collector)
	hostFromUrl(it.CloudLinking, collector)
	hostFromUrl(it.Conda, collector)
	hostFromUrl(it.Docs, collector)
	hostFromUrl(it.Downloads, collector)
	hostFromUrl(it.Issues, collector)
	hostFromUrl(it.Portal, collector)
	hostFromUrl(it.Pypi, collector)
	hostFromUrl(it.PypiFiles, collector)
	hostFromUrl(it.PypiTrusted, collector)
	hostFromUrl(it.RobotPull, collector)
	hostFromUrl(it.Telemetry, collector)
	result := make([]string, 0, len(collector))
	for key, _ := range collector {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
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
