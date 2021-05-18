package operations

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/robocorp/rcc/cloud"
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/pretty"
	"github.com/robocorp/rcc/robot"
	"github.com/robocorp/rcc/settings"
	"github.com/robocorp/rcc/xviper"
	"gopkg.in/yaml.v1"
)

const (
	canaryUrl     = `/canary.txt`
	statusOk      = `ok`
	statusWarning = `warning`
	statusFail    = `fail`
	statusFatal   = `fatal`
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
	result.Details["ROBOCORP_HOME"] = common.RobocorpHome()
	result.Details["ROBOCORP_OVERRIDE_SYSTEM_REQUIREMENTS"] = fmt.Sprintf("%v", common.OverrideSystemRequirements())
	result.Details["RCC_VERBOSE_ENVIRONMENT_BUILDING"] = fmt.Sprintf("%v", common.VerboseEnvironmentBuilding())
	result.Details["user-cache-dir"] = justText(os.UserCacheDir)
	result.Details["user-config-dir"] = justText(os.UserConfigDir)
	result.Details["user-home-dir"] = justText(os.UserHomeDir)
	result.Details["working-dir"] = justText(os.Getwd)
	result.Details["tempdir"] = os.TempDir()
	result.Details["controller"] = common.ControllerIdentity()
	result.Details["user-agent"] = common.UserAgent()
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
	if !common.OverrideSystemRequirements() {
		result.Checks = append(result.Checks, longPathSupportCheck())
	}
	for _, host := range settings.Global.Hostnames() {
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
	supportLongPathUrl := settings.Global.DocsLink("troubleshooting/windows-long-path")
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
	supportGeneralUrl := settings.Global.DocsLink("troubleshooting")
	if !conda.ValidLocation(common.RobocorpHome()) {
		return &common.DiagnosticCheck{
			Type:    "RPA",
			Status:  statusFatal,
			Message: fmt.Sprintf("ROBOCORP_HOME (%s) contains characters that makes RPA fail.", common.RobocorpHome()),
			Link:    supportGeneralUrl,
		}
	}
	return &common.DiagnosticCheck{
		Type:    "RPA",
		Status:  statusOk,
		Message: fmt.Sprintf("ROBOCORP_HOME (%s) is good enough.", common.RobocorpHome()),
		Link:    supportGeneralUrl,
	}
}

func dnsLookupCheck(site string) *common.DiagnosticCheck {
	supportNetworkUrl := settings.Global.DocsLink("troubleshooting/firewall-and-proxies")
	found, err := net.LookupHost(site)
	if err != nil {
		return &common.DiagnosticCheck{
			Type:    "network",
			Status:  statusFail,
			Message: fmt.Sprintf("DNS lookup %q failed: %v", site, err),
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
	supportNetworkUrl := settings.Global.DocsLink("troubleshooting/firewall-and-proxies")
	client, err := cloud.NewClient(settings.Global.DownloadsLink(""))
	if err != nil {
		return &common.DiagnosticCheck{
			Type:    "network",
			Status:  statusFail,
			Message: fmt.Sprintf("%v: %v", settings.Global.DownloadsLink(""), err),
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
		Message: fmt.Sprintf("Canary download successful: %s", settings.Global.DownloadsLink(canaryUrl)),
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
		fmt.Fprintf(sink, " - %-38s...  %q\n", key, value)
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

func ProduceDiagnostics(filename, robotfile string, json bool) (*common.DiagnosticStatus, error) {
	file, err := fileIt(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	result := RunDiagnostics()
	if len(robotfile) > 0 {
		addRobotDiagnostics(robotfile, result)
	}
	settings.Global.Diagnostics(result)
	if json {
		jsonDiagnostics(file, result)
	} else {
		humaneDiagnostics(file, result)
	}
	return result, nil
}

type Unmarshaler func([]byte, interface{}) error

func diagnoseFilesUnmarshal(tool Unmarshaler, label, rootdir string, paths []string, target *common.DiagnosticStatus) {
	supportGeneralUrl := settings.Global.DocsLink("troubleshooting")
	target.Details[fmt.Sprintf("%s-file-count", strings.ToLower(label))] = fmt.Sprintf("%d file(s)", len(paths))
	diagnose := target.Diagnose(label)
	var canary interface{}
	success := true
	investigated := false
	for _, tail := range paths {
		investigated = true
		fullpath := filepath.Join(rootdir, tail)
		content, err := ioutil.ReadFile(fullpath)
		if err != nil {
			diagnose.Fail(supportGeneralUrl, "Problem reading %s file %q: %v", label, tail, err)
			success = false
			continue
		}
		err = tool(content, &canary)
		if err != nil {
			diagnose.Fail(supportGeneralUrl, "Problem parsing %s file %q: %v", label, tail, err)
			success = false
		}
	}
	if investigated && success {
		diagnose.Ok("%s files are readable and can be parsed.", label)
	}
}

func addFileDiagnostics(rootdir string, target *common.DiagnosticStatus) {
	jsons := pathlib.Glob(rootdir, "*.json")
	diagnoseFilesUnmarshal(json.Unmarshal, "JSON", rootdir, jsons, target)
	yamls := pathlib.Glob(rootdir, "*.yaml")
	yamls = append(yamls, pathlib.Glob(rootdir, "*.yml")...)
	diagnoseFilesUnmarshal(yaml.Unmarshal, "YAML", rootdir, yamls, target)
}

func addRobotDiagnostics(robotfile string, target *common.DiagnosticStatus) {
	supportGeneralUrl := settings.Global.DocsLink("troubleshooting")
	config, err := robot.LoadRobotYaml(robotfile, false)
	diagnose := target.Diagnose("Robot")
	if err != nil {
		diagnose.Fail(supportGeneralUrl, "About robot.yaml: %v", err)
	} else {
		config.Diagnostics(target)
	}
	addFileDiagnostics(filepath.Dir(robotfile), target)
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
