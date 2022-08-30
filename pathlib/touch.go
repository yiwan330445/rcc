package pathlib

import (
	"os"
	"time"
)

func TouchWhen(location string, when time.Time) {
	os.Chtimes(location, when, when)
}

func ForceTouchWhen(location string, when time.Time) {
	if !Exists(location) {
		os.WriteFile(location, []byte{}, 0o644)
	}
	TouchWhen(location, when)
}
