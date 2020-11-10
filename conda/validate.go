package conda

import (
	"regexp"
	"strings"

	"github.com/robocorp/rcc/common"
)

var (
	validPathCharacters = regexp.MustCompile("(?i)^[.a-z0-9_:/\\\\]+$")
)

func validateLocations(checked map[string]string) bool {
	success := true
	for name, value := range checked {
		if len(value) == 0 {
			continue
		}
		if strings.ContainsAny(value, " \t") {
			success = false
			common.Log("WARNING!  %s contain spaces. Cannot install miniconda at %v.", name, value)
		}
		if !validPathCharacters.MatchString(value) {
			success = false
			common.Log("WARNING!  %s contain illegal characters. Cannot install miniconda at %v.", name, value)
		}
	}
	if !success {
		common.Log("ERROR!  Cannot install miniconda on your system. See above.")
	}
	return success
}
