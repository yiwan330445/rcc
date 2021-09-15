package conda

import (
	"os"
	"path/filepath"
	"time"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/pretty"
)

func safeRemove(hint, pathling string) error {
	var err error
	if !pathlib.Exists(pathling) {
		common.Debug("[%s] Missing %v, not need to remove.", hint, pathling)
		return nil
	}
	if pathlib.IsDir(pathling) {
		err = renameRemove(pathling)
	} else {
		err = os.Remove(pathling)
	}
	if err != nil {
		pretty.Warning("[%s] %s -> %v", hint, pathling, err)
		pretty.Warning("Make sure that you have rights to %q, and that nothing has locks in there.", pathling)
	} else {
		common.Debug("[%s] Removed %v.", hint, pathling)
	}
	return err
}

func doCleanup(fullpath string, dryrun bool) error {
	if !pathlib.Exists(fullpath) {
		return nil
	}
	if dryrun {
		common.Log("Would be removing: %s", fullpath)
		return nil
	}
	return safeRemove("path", fullpath)
}

func alwaysCleanup(dryrun bool) {
	base := filepath.Join(common.RobocorpHome(), "base")
	live := filepath.Join(common.RobocorpHome(), "live")
	miniconda3 := filepath.Join(common.RobocorpHome(), "miniconda3")
	if dryrun {
		common.Log("Would be removing:")
		common.Log("- %v", base)
		common.Log("- %v", live)
		common.Log("- %v", miniconda3)
		return
	}
	safeRemove("legacy", base)
	safeRemove("legacy", live)
	safeRemove("legacy", miniconda3)
}

func quickCleanup(dryrun bool) error {
	if dryrun {
		common.Log("- %v", common.PipCache())
		common.Log("- %v", common.HolotreeLocation())
		common.Log("- %v", common.RobocorpTempRoot())
		return nil
	}
	safeRemove("cache", common.PipCache())
	err := safeRemove("cache", common.HolotreeLocation())
	if err != nil {
		return err
	}
	return safeRemove("temp", common.RobocorpTempRoot())
}

func spotlessCleanup(dryrun bool) error {
	err := quickCleanup(dryrun)
	if err != nil {
		return err
	}
	if dryrun {
		common.Log("- %v", common.MambaPackages())
		common.Log("- %v", BinMicromamba())
		common.Log("- %v", common.HololibLocation())
		return nil
	}
	safeRemove("cache", common.MambaPackages())
	safeRemove("executable", BinMicromamba())
	return safeRemove("cache", common.HololibLocation())
}

func cleanupTemp(deadline time.Time, dryrun bool) error {
	basedir := common.RobocorpTempRoot()
	handle, err := os.Open(basedir)
	if err != nil {
		return err
	}
	entries, err := handle.Readdir(-1)
	handle.Close()
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.ModTime().After(deadline) {
			continue
		}
		fullpath := filepath.Join(basedir, entry.Name())
		if dryrun {
			common.Log("Would remove temp %v.", fullpath)
			continue
		}
		if entry.IsDir() {
			err = os.RemoveAll(fullpath)
			if err != nil {
				common.Log("Warning[%q]: %v", fullpath, err)
			}
		} else {
			os.Remove(fullpath)
		}
		common.Debug("Removed %v.", fullpath)
	}
	return nil
}

func Cleanup(daylimit int, dryrun, quick, all, micromamba bool) error {
	lockfile := common.RobocorpLock()
	locker, err := pathlib.Locker(lockfile, 30000)
	if err != nil {
		common.Log("Could not get lock on live environment. Quitting!")
		return err
	}
	defer locker.Release()

	alwaysCleanup(dryrun)

	if quick {
		return quickCleanup(dryrun)
	}

	if all {
		return spotlessCleanup(dryrun)
	}

	deadline := time.Now().Add(-48 * time.Duration(daylimit) * time.Hour)
	cleanupTemp(deadline, dryrun)

	if micromamba && err == nil {
		err = doCleanup(common.MambaPackages(), dryrun)
	}
	if micromamba && err == nil {
		err = doCleanup(BinMicromamba(), dryrun)
	}
	return err
}
