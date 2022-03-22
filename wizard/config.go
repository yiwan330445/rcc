package wizard

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/settings"
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
	answers := make(answers)

	warning(len(arguments) > 1, "You provided more than one argument, but only the first one will be\nused as the name.")

	filename, err := ask("Path to (otional) settings.yaml", "", optionalFileValidation("Value should be valid file in filesystem."))
	if err != nil {
		return err
	}
	if len(filename) > 0 {
		settings.TemporalSettingsLayer(filename)
	}

	answers["profile-name"] = firstOf(arguments, settings.Global.Name())
	answers["profile-description"] = settings.Global.Description()
	answers["https-proxy"] = settings.Global.HttpsProxy()
	answers["http-proxy"] = settings.Global.HttpProxy()

	err = questionaire(questions{
		{"profile-name", "Give profile a name", regexpValidation(namePattern, "Use just normal english word characters and no spaces!")},
		{"profile-description", "Give a short description of this profile", regexpValidation(anyPattern, "Description cannot be empty!")},
		{"https-proxy", "URL for https proxy", regexpValidation(proxyPattern, "Must be empty or start with 'http' and should not contain spaces!")},
		{"http-proxy", "URL for http proxy", regexpValidation(proxyPattern, "Must be empty or start with 'http' and should not contain spaces!")},
		{"micromamba-rc", "Path to micromambarc file", optionalFileValidation("Value should be valid file in filesystem.")},
		{"pip-rc", "Path to piprc/pip.ini file", optionalFileValidation("Value should be valid file in filesystem.")},
	}, answers)
	if err != nil {
		return err
	}

	name := answers["profile-name"]
	profile := &settings.Profile{
		Name:        name,
		Description: answers["profile-description"],
	}

	blob, ok := pullFile(answers["settings-yaml"])
	if ok {
		profile.Settings, _ = settings.FromBytes(blob)
	} else {
		profile.Settings = settings.Empty()
	}

	profile.Settings.Network = &settings.Network{
		HttpsProxy: answers["https-proxy"],
		HttpProxy:  answers["http-proxy"],
	}

	profile.Settings.Meta.Name = name
	profile.Settings.Meta.Description = answers["profile-description"]

	blob, ok = pullFile(answers["micromamba-rc"])
	if ok {
		profile.MicroMambaRc = string(blob)
	}

	blob, ok = pullFile(answers["pip-rc"])
	if ok {
		profile.PipRc = string(blob)
	}

	// FIXME: following is just temporary "work in progress" save
	profile.SaveAs(fmt.Sprintf("profile_%s.yaml", strings.ToLower(name)))

	return nil
}

func pullFile(filename string) ([]byte, bool) {
	if !pathlib.IsFile(filename) {
		return nil, false
	}
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, false
	}
	return body, true
}
