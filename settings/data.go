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
	Certificates *Certificates `yaml:"certificates" json:"certificates"`
	Endpoints    *Endpoints    `yaml:"endpoints" json:"endpoints"`
	Hosts        []string      `yaml:"diagnostics-hosts" json:"diagnostics-hosts"`
	Meta         *Meta         `yaml:"meta" json:"meta"`
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

func (it *Settings) Hostnames() []string {
	collector := make(map[string]bool)
	if it.Endpoints != nil {
		for _, name := range it.Endpoints.Hostnames() {
			collector[name] = true
		}
	}
	if it.Hosts != nil {
		for _, name := range it.Hosts {
			collector[name] = true
		}
	}
	result := make([]string, 0, len(collector))
	for key, _ := range collector {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
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

func diagnoseOptionalUrl(link, label string, diagnose common.Diagnoser, correct bool) bool {
	if len(strings.TrimSpace(link)) == 0 {
		return correct
	} else {
		return diagnoseUrl(link, label, diagnose, correct)
	}
}

func (it *Settings) CriticalEnvironmentDiagnostics(target *common.DiagnosticStatus) {
	diagnose := target.Diagnose("settings.yaml")
	correct := true
	if it.Endpoints == nil {
		diagnose.Fatal("", "endpoints section is totally missing")
		correct = false
	} else {
		correct = diagnoseUrl(it.Endpoints.CloudApi, "endpoints/cloud-api", diagnose, correct)
		correct = diagnoseUrl(it.Endpoints.Downloads, "endpoints/downloads", diagnose, correct)
	}
	if correct {
		diagnose.Ok("Toplevel settings are ok.")
	}
}

func (it *Settings) Diagnostics(target *common.DiagnosticStatus) {
	diagnose := target.Diagnose("Settings")
	correct := true
	if it.Certificates == nil {
		diagnose.Warning("", "settings.yaml: certificates section is totally missing")
		correct = false
	}
	if it.Endpoints == nil {
		diagnose.Warning("", "settings.yaml: endpoints section is totally missing")
		correct = false
	} else {
		correct = diagnoseUrl(it.Endpoints.CloudApi, "endpoints/cloud-api", diagnose, correct)
		correct = diagnoseUrl(it.Endpoints.Downloads, "endpoints/downloads", diagnose, correct)

		correct = diagnoseOptionalUrl(it.Endpoints.CloudUi, "endpoints/cloud-ui", diagnose, correct)
		correct = diagnoseOptionalUrl(it.Endpoints.CloudLinking, "endpoints/cloud-linking", diagnose, correct)
		correct = diagnoseOptionalUrl(it.Endpoints.Issues, "endpoints/issues", diagnose, correct)
		correct = diagnoseOptionalUrl(it.Endpoints.Telemetry, "endpoints/telemetry", diagnose, correct)
		correct = diagnoseOptionalUrl(it.Endpoints.Docs, "endpoints/docs", diagnose, correct)
		correct = diagnoseOptionalUrl(it.Endpoints.Conda, "endpoints/conda", diagnose, correct)
		correct = diagnoseOptionalUrl(it.Endpoints.Pypi, "endpoints/pypi", diagnose, correct)
		correct = diagnoseOptionalUrl(it.Endpoints.PypiTrusted, "endpoints/pypi-trusted", diagnose, correct)
	}
	if it.Meta == nil {
		diagnose.Warning("", "settings.yaml: meta section is totally missing")
		correct = false
	}
	if correct {
		diagnose.Ok("Toplevel settings are ok.")
	}
}

type Certificates struct {
	VerifySsl bool `yaml:"verify-ssl" json:"verify-ssl"`
}

type Endpoints struct {
	CloudApi     string `yaml:"cloud-api" json:"cloud-api"`
	CloudLinking string `yaml:"cloud-linking" json:"cloud-linking"`
	CloudUi      string `yaml:"cloud-ui" json:"cloud-ui"`
	Conda        string `yaml:"conda" json:"conda"`
	Docs         string `yaml:"docs" json:"docs"`
	Downloads    string `yaml:"downloads" json:"downloads"`
	Issues       string `yaml:"issues" json:"issues"`
	Pypi         string `yaml:"pypi" json:"pypi"`
	PypiTrusted  string `yaml:"pypi-trusted" json:"pypi-trusted"`
	Telemetry    string `yaml:"telemetry" json:"telemetry"`
}

func justHostAndPort(link string) string {
	if len(link) == 0 {
		return ""
	}
	parsed, err := url.Parse(link)
	if err != nil {
		return ""
	}
	return parsed.Host
}

func hostFromUrl(link string, collector map[string]bool) {
	host := justHostAndPort(link)
	if len(host) > 0 {
		parts := strings.SplitN(host, ":", 2)
		collector[parts[0]] = true
	}
}

func (it *Endpoints) Hostnames() []string {
	collector := make(map[string]bool)
	hostFromUrl(it.CloudApi, collector)
	hostFromUrl(it.CloudLinking, collector)
	hostFromUrl(it.CloudUi, collector)
	hostFromUrl(it.Conda, collector)
	hostFromUrl(it.Docs, collector)
	hostFromUrl(it.Downloads, collector)
	hostFromUrl(it.Issues, collector)
	hostFromUrl(it.Pypi, collector)
	hostFromUrl(it.PypiTrusted, collector)
	hostFromUrl(it.Telemetry, collector)
	result := make([]string, 0, len(collector))
	for key, _ := range collector {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

type Meta struct {
	Source  string `yaml:"source" json:"source"`
	Version string `yaml:"version" json:"version"`
}

type Network struct {
	HttpsProxy string `yaml:"https-proxy" json:"https-proxy"`
	HttpProxy  string `yaml:"http-proxy" json:"http-proxy"`
}

type Profile struct {
	Name        string    `yaml:"name" json:"name"`
	Description string    `yaml:"description" json:"description"`
	Settings    *Settings `yaml:"settings,omitempty" json:"settings,omitempty"`
	Network     *Network  `yaml:"network,omitempty" json:"network,omitempty"`
	Piprc       string    `yaml:"piprc,omitempty" json:"piprc,omitempty"`
	Condarc     string    `yaml:"condarc,omitempty" json:"condarc,omitempty"`
}
