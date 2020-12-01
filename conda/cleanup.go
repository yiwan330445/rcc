package conda

import (
	"os"
	"time"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/pathlib"
)

func Cleanup(daylimit int, dryrun, all bool) error {
	lockfile := MinicondaLock()
	locker, err := pathlib.Locker(lockfile, 30000)
	if err != nil {
		common.Log("Could not get lock on miniconda. Quitting!")
		return err
	}
	defer locker.Release()
	defer os.Remove(lockfile)

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
	if all {
		err = os.RemoveAll(TemplateLocation())
		if err != nil {
			return err
		}
		err = os.RemoveAll(LiveLocation())
		if err != nil {
			return err
		}
	}
	return nil
}
