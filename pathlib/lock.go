package pathlib

import (
	"os"

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
	if common.Trace {
		common.Log("LOCKER: lockless mode release.")
	}
	return nil
}

func Fake() Releaser {
	if common.Trace {
		common.Log("LOCKER: lockless mode.")
	}
	return fake(true)
}
