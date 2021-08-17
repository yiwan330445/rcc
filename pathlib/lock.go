package pathlib

import (
	"os"
	"time"

	"github.com/robocorp/rcc/common"
)

type Releaser interface {
	Release() error
}

type Locked struct {
	*os.File
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
			common.Log("#%d: %s (lock wait)", counter, message)
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
