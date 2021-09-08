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
	return os.RemoveAll(fullpath)
}

func orphanCleanup(dryrun bool) error {
	orphans := OrphanList()
	if len(orphans) == 0 {
		return nil
	}
	if dryrun {
		common.Log("Would be removing orphans:")
		for _, orphan := range orphans {
			common.Log("- %v", orphan)
		}
		return nil
	}
	for _, orphan := range orphans {
		safeRemove("orphan", orphan)
	}
	return nil
}

func quickCleanup(dryrun bool) error {
	if dryrun {
		common.Log("Would be removing:")
		common.Log("- %v", common.BaseLocation())
		common.Log("- %v", common.LiveLocation())
		common.Log("- %v", MinicondaLocation())
		common.Log("- %v", common.PipCache())
		common.Log("- %v", common.HolotreeLocation())
		common.Log("- %v", RobocorpTempRoot())
		return nil
	}
	safeRemove("cache", common.BaseLocation())
	safeRemove("cache", common.LiveLocation())
	safeRemove("cache", MinicondaLocation())
	safeRemove("cache", common.PipCache())
	err := safeRemove("cache", common.HolotreeLocation())
	if err != nil {
		return err
	}
	return safeRemove("temp", RobocorpTempRoot())
}

func spotlessCleanup(dryrun bool) error {
	err := quickCleanup(dryrun)
	if err != nil {
		return err
	}
	if dryrun {
		common.Log("- %v", MambaPackages())
		common.Log("- %v", BinMicromamba())
		common.Log("- %v", common.HololibLocation())
		return nil
	}
	safeRemove("cache", MambaPackages())
	safeRemove("executable", BinMicromamba())
	return safeRemove("cache", common.HololibLocation())
}

func cleanupTemp(deadline time.Time, dryrun bool) error {
	basedir := RobocorpTempRoot()
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

func Cleanup(daylimit int, dryrun, orphans, quick, all, miniconda, micromamba bool) error {
	lockfile := common.RobocorpLock()
	locker, err := pathlib.Locker(lockfile, 30000)
	if err != nil {
		common.Log("Could not get lock on live environment. Quitting!")
		return err
	}
	defer locker.Release()

	if quick {
		return quickCleanup(dryrun)
	}

	if all {
		return spotlessCleanup(dryrun)
	}

	deadline := time.Now().Add(-24 * time.Duration(daylimit) * time.Hour)
	cleanupTemp(deadline, dryrun)
	for _, template := range TemplateList() {
		whenLive, err := LastUsed(LiveFrom(template))
		if err != nil {
			return err
		}
		if !all && whenLive.After(deadline) {
			continue
		}
		whenBase, err := LastUsed(TemplateFrom(template))
		if err != nil {
			return err
		}
		if !all && whenBase.After(deadline) {
			continue
		}
		if dryrun {
			common.Log("Would be removing %v.", template)
			continue
		}
		// FIXME: remove this when base/live removal is done
		RemoveEnvironment(template)
		common.Debug("Removed environment %v.", template)
	}
	if orphans {
		err = orphanCleanup(dryrun)
	}
	if miniconda && err == nil {
		err = doCleanup(MinicondaLocation(), dryrun)
	}
	if micromamba && err == nil {
		err = doCleanup(MambaPackages(), dryrun)
	}
	if micromamba && err == nil {
		err = doCleanup(BinMicromamba(), dryrun)
	}
	return err
}
