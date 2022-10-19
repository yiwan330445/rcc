package pathlib

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/robocorp/rcc/common"
)

type Releaser interface {
	Release() error
}

type Locked struct {
	*os.File
	Marker string
}

type fake bool

func (it fake) Release() error {
	return common.Trace("LOCKER: lockless mode release.")
}

func Fake() Releaser {
	common.Trace("LOCKER: lockless mode.")
	return fake(true)
}

func waitingLockNotification(message string, latch chan bool) {
	delay := 5 * time.Second
	counter := 0
	for {
		select {
		case <-latch:
			return
		case <-time.After(delay):
			counter += 1
			delay *= 3
			common.Log("#%d: %s (rcc lock wait warning)", counter, message)
			common.Timeline("waiting for lock")
		}
	}
}

func LockWaitMessage(message string) func() {
	latch := make(chan bool)
	go waitingLockNotification(message, latch)
	return func() {
		latch <- true
	}
}

func lockPidFilename(lockfile string) string {
	now := time.Now().Format("20060102150405")
	base := filepath.Base(lockfile)
	username := "unspecified"
	who, err := user.Current()
	if err == nil {
		username = who.Username
	}
	marker := fmt.Sprintf("%s.%s.%s.%s.%d.%s", now, username, common.ControllerType, common.HolotreeSpace, os.Getpid(), base)
	return filepath.Join(common.HololibPids(), marker)
}
