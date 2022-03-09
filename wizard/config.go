package wizard

import (
	"fmt"
	"regexp"

	"github.com/robocorp/rcc/common"
)

var (
	proxyPattern = regexp.MustCompile("^(?:http[\\S]*)?$")
	anyPattern   = regexp.MustCompile("^\\s*\\S+")
)

func Configure(arguments []string) error {
	common.Stdout("\n")

	note("You are now configuring a profile to be used in Robocorp toolchain.\n")

	warning(len(arguments) > 1, "You provided more than one argument, but only the first one will be\nused as the name.")
	profileName, err := ask("Give profile a name", firstOf(arguments, "company"), namePattern, "Use just normal english word characters and no spaces!")

	if err != nil {
		return err
	}

	description, err := ask("Give a short description of this profile", "undocumented", anyPattern, "Description cannot be empty!")

	if err != nil {
		return err
	}

	httpsProxy, err := ask("URL for https proxy", "", proxyPattern, "Must be empty or start with 'http' and should not contain spaces!")

	if err != nil {
		return err
	}

	httpProxy, err := ask("URL for http proxy", httpsProxy, proxyPattern, "Must be empty or start with 'http' and should not contain spaces!")

	if err != nil {
		return err
	}

	fmt.Sprintf("%s%v%v%v", description, profileName, httpsProxy, httpProxy)
	return nil
}
