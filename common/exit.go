package common

import (
	"fmt"
)

type ExitCode struct {
	Code    int
	Message string
}

func (it ExitCode) ShowMessage() {
	Log(it.Message)
}

func Exit(code int, format string, rest ...interface{}) {
	panic(ExitCode{
		Code:    code,
		Message: fmt.Sprintf(format, rest...),
	})
}
