// +build windows !darwin !linux

package pretty

import "syscall"

const (
	ENABLE_VIRTUAL_TERMINAL_PROCESSING = 0x400
)

func init() {
	Disabled = true
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	if kernel32 == nil {
		return
	}
	setConsoleMode := kernel32.NewProc("SetConsoleMode")
	if setConsoleMode == nil {
		return
	}
	_, _, err := setConsoleMode.Call(ENABLE_VIRTUAL_TERMINAL_PROCESSING)
	Disabled = err != nil
}
