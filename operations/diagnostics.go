package operations

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/user"
	"runtime"
	"sort"
	"time"

	"github.com/robocorp/rcc/cloud"
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/pretty"
	"github.com/robocorp/rcc/robot"
	"github.com/robocorp/rcc/xviper"
)

const (
	canaryHost         = `https://downloads.robocorp.com`
	canaryUrl          = `/canary.txt`
	supportLongPathUrl = `https://robocorp.com/docs/troubleshooting/windows-long-path`
	supportNetworkUrl  = `https://robocorp.com/docs/troubleshooting/firewall-and-proxies`
	supportGeneralUrl  = `https://robocorp.com/docs/troubleshooting`
	statusOk           = `ok`
	statusWarning      = `warning`
	statusFail         = `fail`
	statusFatal        = `fatal`
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

type stringerr func() (string, error)

func justText(source stringerr) string {
	result, _ := source()
	return result
}

func RunDiagnostics() *common.DiagnosticStatus {
	result := &common.DiagnosticStatus{
		Details: make(map[string]string),
		Checks:  []*common.DiagnosticCheck{},
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
	result.Details["working-dir"] = justText(os.Getwd)
	result.Details["tempdir"] = os.TempDir()
	result.Details["controller"] = common.ControllerIdentity()
	result.Details["installationId"] = xviper.TrackingIdentity()
	result.Details["os"] = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)
	result.Details["cpus"] = fmt.Sprintf("%d", runtime.NumCPU())
	result.Details["when"] = time.Now().Format(time.RFC3339 + " (MST)")

	who, err := user.Current()
	if err == nil {
		result.Details["uid:gid"] = fmt.Sprintf("%s:%s", who.Uid, who.Gid)
	}

	// checks
	result.Checks = append(result.Checks, robocorpHomeCheck())
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

func longPathSupportCheck() *common.DiagnosticCheck {
	if conda.HasLongPathSupport() {
		return &common.DiagnosticCheck{
			Type:    "OS",
			Status:  statusOk,
			Message: "Supports long enough paths.",
			Link:    supportLongPathUrl,
		}
	}
	return &common.DiagnosticCheck{
		Type:    "OS",
		Status:  statusFail,
		Message: "Does not support long path names!",
		Link:    supportLongPathUrl,
	}
}

func robocorpHomeCheck() *common.DiagnosticCheck {
	if !conda.ValidLocation(conda.RobocorpHome()) {
		return &common.DiagnosticCheck{
			Type:    "RPA",
			Status:  statusFatal,
			Message: fmt.Sprintf("ROBOCORP_HOME (%s) contains characters that makes RPA fail.", conda.RobocorpHome()),
			Link:    supportGeneralUrl,
		}
	}
	return &common.DiagnosticCheck{
		Type:    "RPA",
		Status:  statusOk,
		Message: fmt.Sprintf("ROBOCORP_HOME (%s) is good enough.", conda.RobocorpHome()),
		Link:    supportGeneralUrl,
	}
}

func dnsLookupCheck(site string) *common.DiagnosticCheck {
	found, err := net.LookupHost(site)
	if err != nil {
		return &common.DiagnosticCheck{
			Type:    "network",
			Status:  statusFail,
			Message: fmt.Sprintf("DNS lookup %s failed: %v", site, err),
			Link:    supportNetworkUrl,
		}
	}
	return &common.DiagnosticCheck{
		Type:    "network",
		Status:  statusOk,
		Message: fmt.Sprintf("%s found: %v", site, found),
		Link:    supportNetworkUrl,
	}
}

func canaryDownloadCheck() *common.DiagnosticCheck {
	client, err := cloud.NewClient(canaryHost)
	if err != nil {
		return &common.DiagnosticCheck{
			Type:    "network",
			Status:  statusFail,
			Message: fmt.Sprintf("%v: %v", canaryHost, err),
			Link:    supportNetworkUrl,
		}
	}
	request := client.NewRequest(canaryUrl)
	response := client.Get(request)
	if response.Status != 200 || string(response.Body) != "Used to testing connections" {
		return &common.DiagnosticCheck{
			Type:    "network",
			Status:  statusFail,
			Message: fmt.Sprintf("Canary download failed: %d: %s", response.Status, response.Body),
			Link:    supportNetworkUrl,
		}
	}
	return &common.DiagnosticCheck{
		Type:    "network",
		Status:  statusOk,
		Message: fmt.Sprintf("Canary download successful: %s%s", canaryHost, canaryUrl),
		Link:    supportNetworkUrl,
	}
}

func jsonDiagnostics(sink io.Writer, details *common.DiagnosticStatus) {
	form, err := details.AsJson()
	if err != nil {
		pretty.Exit(1, "Error: %s", err)
	}
	fmt.Fprintln(sink, form)
}

func humaneDiagnostics(sink io.Writer, details *common.DiagnosticStatus) {
	fmt.Fprintln(sink, "Diagnostics:")
	keys := make([]string, 0, len(details.Details))
	for key, _ := range details.Details {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := details.Details[key]
		fmt.Fprintf(sink, " - %-25s...  %q\n", key, value)
	}
	fmt.Fprintln(sink, "")
	fmt.Fprintln(sink, "Checks:")
	for _, check := range details.Checks {
		fmt.Fprintf(sink, " - %-8s %-8s %s\n", check.Type, check.Status, check.Message)
	}
}

func fileIt(filename string) (io.WriteCloser, error) {
	if len(filename) == 0 {
		return os.Stdout, nil
	}
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func PrintDiagnostics(filename, robotfile string, json bool) error {
	file, err := fileIt(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	result := RunDiagnostics()
	if len(robotfile) > 0 {
		addRobotDiagnostics(robotfile, result)
	}
	if json {
		jsonDiagnostics(file, result)
	} else {
		humaneDiagnostics(file, result)
	}
	return nil
}

func addRobotDiagnostics(robotfile string, target *common.DiagnosticStatus) {
	config, err := robot.LoadRobotYaml(robotfile, false)
	diagnose := target.Diagnose("Robot")
	if err != nil {
		diagnose.Fail(supportGeneralUrl, "About robot.yaml: %v", err)
	} else {
		config.Diagnostics(target)
	}
}

func RunRobotDiagnostics(robotfile string) *common.DiagnosticStatus {
	result := &common.DiagnosticStatus{
		Details: make(map[string]string),
		Checks:  []*common.DiagnosticCheck{},
	}
	addRobotDiagnostics(robotfile, result)
	return result
}

func PrintRobotDiagnostics(robotfile string, json bool) error {
	result := RunRobotDiagnostics(robotfile)
	if json {
		jsonDiagnostics(os.Stdout, result)
	} else {
		humaneDiagnostics(os.Stderr, result)
	}
	return nil
}
