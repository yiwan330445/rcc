package conda

import (
	"os"
	"path/filepath"
	"time"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/pretty"
)

func safeRemove(hint, pathling string) {
	var err error
	if !pathlib.Exists(pathling) {
		common.Debug("[%s] Missing %v, not need to remove.", hint, pathling)
		return
	}
	if pathlib.IsDir(pathling) {
		err = os.RemoveAll(pathling)
	} else {
		err = os.Remove(pathling)
	}
	if err != nil {
		pretty.Warning("[%s] %s -> %v", hint, pathling, err)
	} else {
		common.Debug("[%s] Removed %v.", hint, pathling)
	}
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
		common.Log("- %v", common.HolotreeLocation())
		common.Log("- %v", common.PipCache())
		common.Log("- %v", RobocorpTempRoot())
		common.Log("- %v", MinicondaLocation())
		return nil
	}
	safeRemove("cache", common.HolotreeLocation())
	safeRemove("cache", common.BaseLocation())
	safeRemove("cache", common.LiveLocation())
	safeRemove("cache", common.PipCache())
	safeRemove("temp", RobocorpTempRoot())
	safeRemove("cache", MinicondaLocation())
	return nil
}

func spotlessCleanup(dryrun bool) error {
	quickCleanup(dryrun)
	if dryrun {
		common.Log("- %v", BinMicromamba())
		common.Log("- %v", common.HololibLocation())
		common.Log("- %v", MambaPackages())
		return nil
	}
	safeRemove("executable", BinMicromamba())
	safeRemove("cache", common.HololibLocation())
	safeRemove("cache", MambaPackages())
	return nil
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
