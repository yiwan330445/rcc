package htfs

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/fail"
)

func RecordCondaEnvironment(condafile string, force bool) (lib Library, err error) {
	defer fail.Around(&err)

	right, err := conda.ReadCondaYaml(condafile)
	fail.On(err != nil, "Could not load environmet config %q, reason: %w", condafile, err)

	content, err := right.AsYaml()
	fail.On(err != nil, "YAML error with %q, reason: %w", condafile, err)

	return RecordEnvironment([]byte(content), force)
}

func RecordEnvironment(blueprint []byte, force bool) (lib Library, err error) {
	defer fail.Around(&err)

	tree, err := New()
	fail.On(err != nil, "Failed to create holotree location, reason %w.", err)

	// following must be setup here
	common.StageFolder = tree.Stage()
	common.Stageonly = true
	common.Liveonly = true

	err = os.RemoveAll(tree.Stage())
	fail.On(err != nil, "Failed to clean stage, reason %w.", err)

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
	return tree, nil
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
