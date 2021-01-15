package conda

import (
	"fmt"
	"os"
	"time"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/pathlib"
)

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
	if anyLeasedEnvironment() {
		return fmt.Errorf("Cannot clean everything, since there are some leased environments!")
	}
	if dryrun {
		common.Log("Would be removing:")
		common.Log("- %v", TemplateLocation())
		common.Log("- %v", LiveLocation())
		common.Log("- %v", PipCache())
		common.Log("- %v", MambaPackages())
		common.Log("- %v", BinMicromamba())
		common.Log("- %v", RobocorpTempRoot())
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
	err = os.RemoveAll(MambaPackages())
	if err != nil {
		return err
	}
	common.Debug("Removed directory %v.", MambaPackages())
	err = os.Remove(BinMicromamba())
	if err != nil {
		return err
	}
	common.Debug("Removed executable %v.", BinMicromamba())
	err = os.RemoveAll(RobocorpTempRoot())
	if err != nil {
		return err
	}
	common.Debug("Removed directory %v.", RobocorpTempRoot())
	return nil
}

func anyLeasedEnvironment() bool {
	for _, template := range TemplateList() {
		if IsLeasedEnvironment(template) {
			return true
		}
	}
	return false
}

func Cleanup(daylimit int, dryrun, orphans, all, miniconda, micromamba bool) error {
	lockfile := RobocorpLock()
	locker, err := pathlib.Locker(lockfile, 30000)
	if err != nil {
		common.Log("Could not get lock on live environment. Quitting!")
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
		if IsLeasedEnvironment(template) {
			common.Log("WARNING: %q is leased by %q and wont be cleaned up!", template, WhoLeased(template))
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
