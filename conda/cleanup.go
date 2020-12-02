package conda

import (
	"os"
	"time"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/pathlib"
)

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
		var err error
		if pathlib.IsDir(orphan) {
			err = os.RemoveAll(orphan)
		} else {
			err = os.Remove(orphan)
		}
		if err != nil {
			return err
		}
		common.Debug("Removed orphan %v.", orphan)
	}
	return nil
}

func spotlessCleanup(dryrun bool) error {
	if dryrun {
		common.Log("Would be removing:")
		common.Log("- %v", TemplateLocation())
		common.Log("- %v", LiveLocation())
		common.Log("- %v", PipCache())
		common.Log("- %v", CondaPackages())
		return nil
	}
	err := os.RemoveAll(TemplateLocation())
	if err != nil {
		return err
	}
	common.Debug("Removed directory %v.", TemplateLocation())
	err = os.RemoveAll(LiveLocation())
	if err != nil {
		return err
	}
	common.Debug("Removed directory %v.", LiveLocation())
	err = os.RemoveAll(PipCache())
	if err != nil {
		return err
	}
	common.Debug("Removed directory %v.", PipCache())
	err = os.RemoveAll(CondaPackages())
	if err != nil {
		return err
	}
	common.Debug("Removed directory %v.", CondaPackages())
	return nil
}

func Cleanup(daylimit int, dryrun, orphans, all bool) error {
	lockfile := MinicondaLock()
	locker, err := pathlib.Locker(lockfile, 30000)
	if err != nil {
		common.Log("Could not get lock on miniconda. Quitting!")
		return err
	}
	defer locker.Release()
	defer os.Remove(lockfile)

	if all {
		return spotlessCleanup(dryrun)
	}

	deadline := time.Now().Add(-24 * time.Duration(daylimit) * time.Hour)
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
		return orphanCleanup(dryrun)
	}
	return nil
}
