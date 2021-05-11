package htfs

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/robocorp/rcc/anywork"
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/fail"
	"github.com/robocorp/rcc/robot"
)

func NewEnvironment(force bool, condafile string) (label string, err error) {
	defer fail.Around(&err)

	common.Timeline("new holotree environment")
	_, holotreeBlueprint, err := ComposeFinalBlueprint([]string{condafile}, "")
	fail.On(err != nil, "%s", err)

	anywork.Scale(200)

	tree, err := New()
	fail.On(err != nil, "%s", err)
	common.Timeline("holotree library created")

	if !tree.HasBlueprint(holotreeBlueprint) && common.Liveonly {
		tree = Virtual()
		common.Timeline("downgraded to virtual holotree library")
	}
	err = RecordEnvironment(tree, holotreeBlueprint, force)
	fail.On(err != nil, "%s", err)

	path, err := tree.Restore(holotreeBlueprint, []byte(common.ControllerIdentity()), []byte(common.HolotreeSpace))
	fail.On(err != nil, "Failed to restore blueprint %q, reason: %v", string(holotreeBlueprint), err)
	return path, nil
}

func RecordCondaEnvironment(tree Library, condafile string, force bool) (err error) {
	defer fail.Around(&err)

	right, err := conda.ReadCondaYaml(condafile)
	fail.On(err != nil, "Could not load environmet config %q, reason: %w", condafile, err)

	content, err := right.AsYaml()
	fail.On(err != nil, "YAML error with %q, reason: %w", condafile, err)

	return RecordEnvironment(tree, []byte(content), force)
}

func RecordEnvironment(tree Library, blueprint []byte, force bool) (err error) {
	defer fail.Around(&err)

	// following must be setup here
	common.StageFolder = tree.Stage()
	common.Stageonly = true
	common.Liveonly = true

	err = os.RemoveAll(tree.Stage())
	fail.On(err != nil, "Failed to clean stage, reason %v.", err)

	err = os.MkdirAll(tree.Stage(), 0o755)
	fail.On(err != nil, "Failed to create stage, reason %v.", err)

	common.Debug("Holotree stage is %q.", tree.Stage())
	exists := tree.HasBlueprint(blueprint)
	common.Debug("Has blueprint environment: %v", exists)

	if force || !exists {
		identityfile := filepath.Join(tree.Stage(), "identity.yaml")
		err = ioutil.WriteFile(identityfile, blueprint, 0o640)
		fail.On(err != nil, "Failed to save %q, reason %w.", identityfile, err)
		label, err := conda.NewEnvironment(force, identityfile)
		fail.On(err != nil, "Failed to create environment, reason %w.", err)
		common.Debug("Label: %q", label)
	}

	if force || !exists {
		err := tree.Record(blueprint)
		fail.On(err != nil, "Failed to record blueprint %q, reason: %w", string(blueprint), err)
	}
	return nil
}

func FindEnvironment(fragment string) []string {
	result := make([]string, 0, 10)
	for directory, _ := range Spacemap() {
		name := filepath.Base(directory)
		if strings.Contains(name, fragment) {
			result = append(result, name)
		}
	}
	return result
}

func RemoveHolotreeSpace(label string) (err error) {
	defer fail.Around(&err)

	for directory, metafile := range Spacemap() {
		name := filepath.Base(directory)
		if name != label {
			continue
		}
		os.Remove(metafile)
		err = os.RemoveAll(directory)
		fail.On(err != nil, "Problem removing %q, reason: %s.", directory, err)
	}
	return nil
}

func RobotBlueprints(userBlueprints []string, packfile string) (robot.Robot, []string) {
	var err error
	var config robot.Robot

	blueprints := make([]string, 0, len(userBlueprints)+2)

	if len(packfile) > 0 {
		config, err = robot.LoadRobotYaml(packfile, false)
		if err == nil {
			blueprints = append(blueprints, config.CondaConfigFile())
		}
	}

	return config, append(blueprints, userBlueprints...)
}

func ComposeFinalBlueprint(userFiles []string, packfile string) (config robot.Robot, blueprint []byte, err error) {
	defer fail.Around(&err)

	var left, right *conda.Environment

	config, filenames := RobotBlueprints(userFiles, packfile)

	for _, filename := range filenames {
		left = right
		right, err = conda.ReadCondaYaml(filename)
		fail.On(err != nil, "Failure: %v", err)
		if left == nil {
			continue
		}
		right, err = left.Merge(right)
		fail.On(err != nil, "Failure: %v", err)
	}
	fail.On(right == nil, "Missing environment specification(s).")
	content, err := right.AsYaml()
	fail.On(err != nil, "YAML error: %v", err)
	return config, []byte(strings.TrimSpace(content)), nil
}
