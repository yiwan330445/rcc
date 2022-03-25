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
	note("If you want to clear some value, try giving just one space as a value.\n")
	note("If you want to use default value, just press enter.\n")

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
	answers["ssl-verify"] = fmt.Sprintf("%v", settings.Global.VerifySsl())
	answers["ssl-no-revoke"] = fmt.Sprintf("%v", settings.Global.NoRevocation())

	err = questionaire(questions{
		{"profile-name", "Give profile a name", regexpValidation(namePattern, "Use just normal english word characters and no spaces!")},
		{"profile-description", "Give a short description of this profile", regexpValidation(anyPattern, "Description cannot be empty!")},
		{"https-proxy", "URL for https proxy", regexpValidation(proxyPattern, "Must be empty or start with 'http' and should not contain spaces!")},
		{"http-proxy", "URL for http proxy", regexpValidation(proxyPattern, "Must be empty or start with 'http' and should not contain spaces!")},
		{"ssl-verify", "Verify SSL certificated (ssl-verify)", memberValidation([]string{"true", "false"}, "Must be either true or false")},
		{"ssl-no-revoke", "Do not check SSL revocations (ssl-no-revoke)", memberValidation([]string{"true", "false"}, "Must be either true or false")},
		{"micromamba-rc", "Optional path to micromambarc file", optionalFileValidation("Value should be valid file in filesystem.")},
		{"pip-rc", "Optional path to piprc/pip.ini file", optionalFileValidation("Value should be valid file in filesystem.")},
		{"ca-bundle", "Optional path to CA bundle [pem format] file", optionalFileValidation("Value should be valid file in filesystem.")},
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

	profile.Settings.Certificates = &settings.Certificates{
		VerifySsl:   answers["ssl-verify"] == "true",
		SslNoRevoke: answers["ssl-no-revoke"] == "true",
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

	blob, ok = pullFile(answers["ca-bundle"])
	if ok {
		profile.CaBundle = string(blob)
	}

	profilename := fmt.Sprintf("profile_%s.yaml", strings.ToLower(name))
	profile.SaveAs(profilename)

	note(fmt.Sprintf("Saved profile into file %q.", profilename))

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
