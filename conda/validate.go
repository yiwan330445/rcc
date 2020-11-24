package conda

import (
	"regexp"
	"strings"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/pretty"
)

const (
	longPathSupportArticle = `https://robocorp.com/docs/product-manuals/robocorp-lab/troubleshooting#windows-has-to-have-long-filenames-support-on`
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
			common.Log("%sWARNING!  %s contain spaces. Cannot install miniconda at %v.%s", pretty.Red, name, value, pretty.Reset)
		}
		if !validPathCharacters.MatchString(value) {
			success = false
			common.Log("%sWARNING!  %s contain illegal characters. Cannot install miniconda at %v.%s", pretty.Red, name, value, pretty.Reset)
		}
	}
	if !success {
		common.Log("%sERROR!  Cannot install miniconda on your system. See above.%s", pretty.Red, pretty.Reset)
	}
	return success
}
