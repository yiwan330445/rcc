package journal

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/fail"
)

var (
	spacePattern = regexp.MustCompile("\\s+")
)

type Event struct {
	When       int64  `json:"when"`
	Controller string `json:"controller"`
	Event      string `json:"event"`
	Detail     string `json:"detail"`
	Comment    string `json:"comment,omitempty"`
}

func Unify(value string) string {
	return strings.TrimSpace(spacePattern.ReplaceAllString(value, " "))
}

func Post(event, detail, commentForm string, fields ...interface{}) (err error) {
	defer fail.Around(&err)
	message := Event{
		When:       common.When,
		Controller: common.ControllerIdentity(),
		Event:      Unify(event),
		Detail:     detail,
		Comment:    Unify(fmt.Sprintf(commentForm, fields...)),
	}
	blob, err := json.Marshal(message)
	fail.On(err != nil, "Could not serialize event: %v -> %v", event, err)
	return appendJournal(blob)
}

func appendJournal(blob []byte) (err error) {
	defer fail.Around(&err)
	handle, err := os.OpenFile(common.EventJournal(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o640)
	fail.On(err != nil, "Failed to open event journal %v -> %v", common.EventJournal(), err)
	defer handle.Close()
	_, err = handle.Write(blob)
	fail.On(err != nil, "Failed to write event journal %v -> %v", common.EventJournal(), err)
	_, err = handle.Write([]byte{'\n'})
	fail.On(err != nil, "Failed to write event journal %v -> %v", common.EventJournal(), err)
	return handle.Sync()
}
