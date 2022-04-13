package pretty

import (
	"fmt"

	"github.com/robocorp/rcc/common"
)

func Ok() error {
	common.Log("%sOK.%s", Green, Reset)
	return nil
}

func Note(format string, rest ...interface{}) {
	niceform := fmt.Sprintf("%s%sNote: %s%s", Cyan, Bold, format, Reset)
	common.Log(niceform, rest...)
}

func Warning(format string, rest ...interface{}) {
	niceform := fmt.Sprintf("%sWarning: %s%s", Yellow, format, Reset)
	common.Log(niceform, rest...)
}

func Exit(code int, format string, rest ...interface{}) {
	var niceform string
	if code == 0 {
		niceform = fmt.Sprintf("%s%s%s", Green, format, Reset)
	} else {
		niceform = fmt.Sprintf("%s%s%s", Red, format, Reset)
	}
	common.Exit(code, niceform, rest...)
}

// Guard watches, that only truthful shall pass. Otherwise exits with code and details.
func Guard(truth bool, code int, format string, rest ...interface{}) {
	if !truth {
		Exit(code, format, rest...)
	}
}
