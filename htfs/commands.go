package htfs

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/fail"
)

func RecordEnvironment(blueprint []byte, force bool) (lib Library, err error) {
	defer fail.Around(&err)

	tree, err := New(common.RobocorpHome())
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
