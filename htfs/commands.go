package htfs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/robocorp/rcc/anywork"
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/fail"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/robot"
)

func Platform() string {
	return strings.ToLower(fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH))
}

func NewEnvironment(force bool, condafile, holozip string) (label string, err error) {
	defer fail.Around(&err)

	callback := pathlib.LockWaitMessage("Serialized environment creation")
	locker, err := pathlib.Locker(common.HolotreeLock(), 30000)
	callback()
	fail.On(err != nil, "Could not get lock for holotree. Quiting.")
	defer locker.Release()

	haszip := len(holozip) > 0

	common.Timeline("new holotree environment")
	_, holotreeBlueprint, err := ComposeFinalBlueprint([]string{condafile}, "")
	fail.On(err != nil, "%s", err)

	anywork.Scale(100)

	tree, err := New()
	fail.On(err != nil, "%s", err)
	common.Timeline("holotree library created")

	if !haszip && !tree.HasBlueprint(holotreeBlueprint) && common.Liveonly {
		tree = Virtual()
		common.Timeline("downgraded to virtual holotree library")
	}
	var library Library
	if haszip {
		library, err = ZipLibrary(holozip)
		fail.On(err != nil, "Failed to load %q -> %s", holozip, err)
		common.Timeline("downgraded to holotree zip library")
	} else {
		err = RecordEnvironment(tree, holotreeBlueprint, force)
		fail.On(err != nil, "%s", err)
		library = tree
	}

	path, err := library.Restore(holotreeBlueprint, []byte(common.ControllerIdentity()), []byte(common.HolotreeSpace))
	fail.On(err != nil, "Failed to restore blueprint %q, reason: %v", string(holotreeBlueprint), err)
	common.EnvironmentHash = BlueprintHash(holotreeBlueprint)
	return path, nil
}

func RecordCondaEnvironment(tree MutableLibrary, condafile string, force bool) (err error) {
	defer fail.Around(&err)

	locker, err := pathlib.Locker(common.HolotreeLock(), 30000)
	fail.On(err != nil, "Could not get lock for holotree. Quiting.")
	defer locker.Release()

	right, err := conda.ReadCondaYaml(condafile)
	fail.On(err != nil, "Could not load environmet config %q, reason: %w", condafile, err)

	content, err := right.AsYaml()
	fail.On(err != nil, "YAML error with %q, reason: %w", condafile, err)

	return RecordEnvironment(tree, []byte(content), force)
}

func CleanupHolotreeStage(tree MutableLibrary) error {
	common.Timeline("holotree stage removal start")
	defer common.Timeline("holotree stage removal done")
	return os.RemoveAll(tree.Stage())
}

func RecordEnvironment(tree MutableLibrary, blueprint []byte, force bool) (err error) {
	defer fail.Around(&err)

	// following must be setup here
	common.StageFolder = tree.Stage()
	common.Stageonly = true
	common.Liveonly = true

	common.Debug("Holotree stage is %q.", tree.Stage())
	exists := tree.HasBlueprint(blueprint)
	common.Debug("Has blueprint environment: %v", exists)

	if force || !exists {
		err = CleanupHolotreeStage(tree)
		fail.On(err != nil, "Failed to clean stage, reason %v.", err)

		err = os.MkdirAll(tree.Stage(), 0o755)
		fail.On(err != nil, "Failed to create stage, reason %v.", err)

		identityfile := filepath.Join(tree.Stage(), "identity.yaml")
		err = ioutil.WriteFile(identityfile, blueprint, 0o644)
		fail.On(err != nil, "Failed to save %q, reason %w.", identityfile, err)
		label, err := conda.NewEnvironment(force, identityfile)
		fail.On(err != nil, "Failed to create environment, reason %w.", err)
		common.Debug("Label: %q", label)

		err = tree.Record(blueprint)
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

func InstallationPlan(hash string) (string, bool) {
	finalplan := filepath.Join(common.HolotreeLocation(), hash, "rcc_plan.log")
	return finalplan, pathlib.IsFile(finalplan)
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
