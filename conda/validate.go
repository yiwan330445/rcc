package conda

import (
	"regexp"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/pretty"
)

const (
	longPathSupportArticle = `https://robocorp.com/docs/product-manuals/robocorp-lab/troubleshooting#windows-has-to-have-long-filenames-support-on`
)

var (
	validPathCharacters = regexp.MustCompile("(?i)^[.a-z0-9_:/\\\\~-]+$")
)

func ValidLocation(value string) bool {
	return validPathCharacters.MatchString(value)
}

func validateLocations(checked map[string]string) bool {
	success := true
	for name, value := range checked {
		if len(value) == 0 {
			continue
		}
		if !ValidLocation(value) {
			success = false
			common.Log("%sWARNING!  %s contain illegal characters. Cannot use tooling with path %q.%s", pretty.Yellow, name, value, pretty.Reset)
		}
	}
	if !success {
		common.Log("%sWARNING!  Python pip might not work correctly in your system. See above.%s", pretty.Yellow, pretty.Reset)
	}
	return success
}

func ValidateLocations() bool {
	checked := map[string]string{
		//"Environment variable 'TMP'":        os.Getenv("TMP"),
		//"Environment variable 'TEMP'":       os.Getenv("TEMP"),
		"Path to 'ROBOCORP_HOME' directory": common.RobocorpHome(),
	}
	// 7.1.2021 -- just warnings for now -- JMP:FIXME:JMP later
	validateLocations(checked)
	return true
	// return validateLocations(checked)
}
