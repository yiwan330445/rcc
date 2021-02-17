package common

import (
	"encoding/json"
	"fmt"
)

const (
	StatusOk      = `ok`
	StatusWarning = `warning`
	StatusFail    = `fail`
	StatusFatal   = `fatal`
)

type Diagnoser func(status, link, form string, details ...interface{})

func (it Diagnoser) Ok(form string, details ...interface{}) {
	it(StatusOk, "", form, details...)
}

func (it Diagnoser) Warning(link, form string, details ...interface{}) {
	it(StatusWarning, link, form, details...)
}

func (it Diagnoser) Fail(link, form string, details ...interface{}) {
	it(StatusFail, link, form, details...)
}

func (it Diagnoser) Fatal(link, form string, details ...interface{}) {
	it(StatusFatal, link, form, details...)
}

type DiagnosticStatus struct {
	Details map[string]string  `json:"details"`
	Checks  []*DiagnosticCheck `json:"checks"`
}

type DiagnosticCheck struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Link    string `json:"url"`
}

func (it *DiagnosticStatus) check(kind, status, message, link string) {
	it.Checks = append(it.Checks, &DiagnosticCheck{kind, status, message, link})
}

func (it *DiagnosticStatus) Diagnose(kind string) Diagnoser {
	return func(status, link, form string, details ...interface{}) {
		it.check(kind, status, fmt.Sprintf(form, details...), link)
	}
}

func (it *DiagnosticStatus) AsJson() (string, error) {
	body, err := json.MarshalIndent(it, "", "  ")
	if err != nil {
		return "", err
	}
	return string(body), nil
}
