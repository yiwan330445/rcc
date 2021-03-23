package common

import (
	"os"
	"path/filepath"
)

const (
	defaultRobocorpLocation = "$HOME/.robocorp"
)

func ExpandPath(entry string) string {
	intermediate := os.ExpandEnv(entry)
	result, err := filepath.Abs(intermediate)
	if err != nil {
		return intermediate
	}
	return result
}
