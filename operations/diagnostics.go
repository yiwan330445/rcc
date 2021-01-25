package operations

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"runtime"

	"github.com/robocorp/rcc/cloud"
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/xviper"
)

const (
	canaryHost = `https://downloads.robocorp.com`
	canaryUrl  = `/canary.txt`
)

var (
	checkedHosts = []string{
		`api.eu1.robocloud.eu`,
		`downloads.robocorp.com`,
		`pypi.org`,
		`conda.anaconda.org`,
		`github.com`,
		`files.pythonhosted.org`,
	}
)

type DiagnosticStatus struct {
	Details map[string]string  `json:"details"`
	Checks  []*DiagnosticCheck `json:"checks"`
}

type DiagnosticCheck struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (it *DiagnosticStatus) AsJson() (string, error) {
	body, err := json.MarshalIndent(it, "", "  ")
	if err != nil {
		return "", err
	}
	return string(body), nil
}

type stringerr func() (string, error)

func justText(source stringerr) string {
	result, _ := source()
	return result
}

func RunDiagnostics() *DiagnosticStatus {
	result := &DiagnosticStatus{
		Details: make(map[string]string),
		Checks:  []*DiagnosticCheck{},
	}
	executable, _ := os.Executable()
	result.Details["executable"] = executable
	result.Details["rcc"] = common.Version
	result.Details["stats"] = rccStatusLine()
	result.Details["micromamba"] = conda.MicromambaVersion()
	result.Details["ROBOCORP_HOME"] = conda.RobocorpHome()
	result.Details["user-cache-dir"] = justText(os.UserCacheDir)
	result.Details["user-config-dir"] = justText(os.UserConfigDir)
	result.Details["user-home-dir"] = justText(os.UserHomeDir)
	//result.Details["hostname"] = justText(os.Hostname)
	result.Details["tempdir"] = os.TempDir()
	result.Details["os"] = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)
	result.Details["cpus"] = fmt.Sprintf("%d", runtime.NumCPU())

	// checks
	result.Checks = append(result.Checks, longPathSupportCheck())
	for _, host := range checkedHosts {
		result.Checks = append(result.Checks, dnsLookupCheck(host))
	}
	result.Checks = append(result.Checks, canaryDownloadCheck())
	return result
}

func rccStatusLine() string {
	requests := xviper.GetInt("stats.env.request")
	hits := xviper.GetInt("stats.env.hit")
	dirty := xviper.GetInt("stats.env.dirty")
	misses := xviper.GetInt("stats.env.miss")
	failures := xviper.GetInt("stats.env.failures")
	merges := xviper.GetInt("stats.env.merges")
	templates := len(conda.TemplateList())
	return fmt.Sprintf("%d environments, %d requests, %d merges, %d hits, %d dirty, %d misses, %d failures | %s", templates, requests, merges, hits, dirty, misses, failures, xviper.TrackingIdentity())
}

func longPathSupportCheck() *DiagnosticCheck {
	if conda.HasLongPathSupport() {
		return &DiagnosticCheck{
			Type:    "OS",
			Status:  "ok",
			Message: "Supports long enough paths.",
		}
	}
	return &DiagnosticCheck{
		Type:    "OS",
		Status:  "fail",
		Message: "Does not support long path names!",
	}
}

func dnsLookupCheck(site string) *DiagnosticCheck {
	found, err := net.LookupHost(site)
	if err != nil {
		return &DiagnosticCheck{
			Type:    "network",
			Status:  "fail",
			Message: fmt.Sprintf("DNS lookup %s failed: %v", site, err),
		}
	}
	return &DiagnosticCheck{
		Type:    "network",
		Status:  "ok",
		Message: fmt.Sprintf("%s found: %v", site, found),
	}
}

func canaryDownloadCheck() *DiagnosticCheck {
	client, err := cloud.NewClient(canaryHost)
	if err != nil {
		return &DiagnosticCheck{
			Type:    "network",
			Status:  "fail",
			Message: fmt.Sprintf("%v: %v", canaryHost, err),
		}
	}
	request := client.NewRequest(canaryUrl)
	response := client.Get(request)
	if response.Status != 200 || string(response.Body) != "Used to testing connections" {
		return &DiagnosticCheck{
			Type:    "network",
			Status:  "fail",
			Message: fmt.Sprintf("Canary download failed: %d: %s", response.Status, response.Body),
		}
	}
	return &DiagnosticCheck{
		Type:    "network",
		Status:  "ok",
		Message: fmt.Sprintf("Canary download successful: %s%s", canaryHost, canaryUrl),
	}
}
