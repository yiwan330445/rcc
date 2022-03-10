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

type question struct {
	Identity  string
	Question  string
	Validator Validator
}

type questions []question

type answers map[string]string

func questionaire(questions questions, answers answers) error {
	for _, question := range questions {
		previous := answers[question.Identity]
		indirect, ok := answers[previous]
		if ok {
			previous = indirect
		}
		answer, err := ask(question.Question, previous, question.Validator)
		if err != nil {
			return err
		}
		answers[question.Identity] = answer
	}
	return nil
}

func Configure(arguments []string) error {
	common.Stdout("\n")

	note("You are now configuring a profile to be used in Robocorp toolchain.\n")

	warning(len(arguments) > 1, "You provided more than one argument, but only the first one will be\nused as the name.")

	answers := make(answers)
	answers["profile-name"] = firstOf(arguments, "company")
	answers["profile-description"] = "undocumented"
	answers["https-proxy"] = ""
	answers["http-proxy"] = "https-proxy"
	answers["settings-yaml"] = ""
	answers["conda-rc"] = ""
	answers["pip-rc"] = ""

	err := questionaire(questions{
		{"profile-name", "Give profile a name", regexpValidation(namePattern, "Use just normal english word characters and no spaces!")},
		{"profile-description", "Give a short description of this profile", regexpValidation(anyPattern, "Description cannot be empty!")},
		{"https-proxy", "URL for https proxy", regexpValidation(proxyPattern, "Must be empty or start with 'http' and should not contain spaces!")},
		{"http-proxy", "URL for http proxy", regexpValidation(proxyPattern, "Must be empty or start with 'http' and should not contain spaces!")},
		{"settings-yaml", "Path to settings.yaml file", optionalFileValidation("Value should be valid file in filesystem.")},
		{"conda-rc", "Path to condarc file", optionalFileValidation("Value should be valid file in filesystem.")},
		{"pip-rc", "Path to piprc/pip.ini file", optionalFileValidation("Value should be valid file in filesystem.")},
	}, answers)

	if err != nil {
		return err
	}

	fmt.Println(answers)

	return nil
}
