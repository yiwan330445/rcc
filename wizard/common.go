package wizard

import (
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/pretty"
)

type WizardFn func([]string) error

func warning(condition bool, message string) {
	if condition {
		common.Stdout("%s%s%s\n\n", pretty.Yellow, message, pretty.Reset)
	}
}

func firstOf(arguments []string, missing string) string {
	if len(arguments) > 0 {
		return arguments[0]
	}
	return missing
}
