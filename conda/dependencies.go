package conda

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/fail"
	"github.com/robocorp/rcc/pretty"
	"gopkg.in/yaml.v2"
)

type dependency struct {
	Name    string `yaml:"name"    json:"name"`
	Version string `yaml:"version" json:"version"`
	Origin  string `yaml:"origin"  json:"channel"`
}

type dependencies []*dependency

func parseDependencies(origin string, output []byte) (dependencies, error) {
	result := make(dependencies, 0, 100)
	err := json.Unmarshal(output, &result)
	if err != nil {
		return nil, err
	}
	if len(origin) == 0 {
		return result, nil
	}
	for _, here := range result {
		if len(here.Origin) == 0 {
			here.Origin = origin
		}
	}
	return result, nil
}

func fillDependencies(context, targetFolder string, seen map[string]string, collector dependencies, command ...string) (_ dependencies, err error) {
	defer fail.Around(&err)

	task, err := livePrepare(targetFolder, command...)
	fail.On(err != nil, "%v", err)
	out, _, err := task.CaptureOutput()
	fail.On(err != nil, "%v", err)
	listing, err := parseDependencies(context, []byte(out))
	fail.On(err != nil, "%v", err)
	for _, entry := range listing {
		found, ok := seen[strings.ToLower(entry.Name)]
		if ok && found == entry.Version {
			continue
		}
		collector = append(collector, entry)
		seen[strings.ToLower(entry.Name)] = entry.Version
	}
	return collector, nil
}

func goldenMasterFilename(targetFolder string) string {
	return filepath.Join(targetFolder, "golden-ee.yaml")
}

func goldenMaster(targetFolder string, pipUsed bool) (err error) {
	defer fail.Around(&err)

	seen := make(map[string]string)
	collector := make(dependencies, 0, 100)
	collector, err = fillDependencies("mamba", targetFolder, seen, collector, BinMicromamba(), "list", "--json")
	fail.On(err != nil, "Failed to list micromamba dependencies, reason: %v", err)
	if pipUsed {
		collector, err = fillDependencies("pypi", targetFolder, seen, collector, "pip", "list", "--isolated", "--local", "--format", "json")
		fail.On(err != nil, "Failed to list pip dependencies, reason: %v", err)
	}
	sort.SliceStable(collector, func(left, right int) bool {
		return strings.ToLower(collector[left].Name) < strings.ToLower(collector[right].Name)
	})
	body, err := yaml.Marshal(collector)
	fail.On(err != nil, "Failed to make yaml, reason: %v", err)
	goldenfile := goldenMasterFilename(targetFolder)
	common.Debug("%sGolden EE file at: %v%s", pretty.Yellow, goldenfile, pretty.Reset)
	return os.WriteFile(goldenfile, body, 0644)
}

func LoadWantedDependencies(filename string) (_ dependencies, err error) {
	defer fail.Around(&err)

	body, err := os.ReadFile(filename)
	fail.On(err != nil, "Failed to read dependencies from %q, reason: %v.", filename, err)
	result := make(dependencies, 0, 100)
	err = yaml.Unmarshal(body, &result)
	fail.On(err != nil, "Failed to parse dependencies from %q, reason: %v.", filename, err)
	return result, nil
}

func DumpEnvironmentDependencies(targetFolder string) (err error) {
	defer fail.Around(&err)

	filename := goldenMasterFilename(targetFolder)
	dependencyList, err := LoadWantedDependencies(filename)
	fail.On(err != nil, "%v", err)

	tabbed := tabwriter.NewWriter(os.Stderr, 2, 4, 2, ' ', 0)
	tabbed.Write([]byte("No.\tPackage\tVersion\tOrigin\n"))
	tabbed.Write([]byte("---\t-------\t-------\t------\n"))
	for at, entry := range dependencyList {
		data := fmt.Sprintf("%3d\t%s\t%s\t%s\n", at+1, entry.Name, entry.Version, entry.Origin)
		tabbed.Write([]byte(data))
	}
	tabbed.Flush()
	return nil
}
