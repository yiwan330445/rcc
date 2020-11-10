package pathlib

import (
	"os"
	"time"
)

func TouchWhen(location string, when time.Time) {
	os.Chtimes(location, when, when)
}
