package cmd

import (
	"path/filepath"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/pretty"
)

func osSpecificHolotreeSharing(enable bool) {
	pathlib.ForceShared()
	parent := filepath.Dir(common.HoloLocation())
	_, err := pathlib.ForceSharedDir(parent)
	pretty.Guard(err == nil, 1, "Could not enable shared location at %q, reason: %v", parent, err)
	_, err = pathlib.ForceSharedDir(common.HoloLocation())
	pretty.Guard(err == nil, 2, "Could not enable shared location at %q, reason: %v", common.HoloLocation(), err)
}
