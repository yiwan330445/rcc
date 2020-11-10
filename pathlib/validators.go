package pathlib

import (
	"errors"
	"fmt"
	"os"
)

func FileExist(name string) bool {
	stat, err := os.Stat(name)
	if err != nil {
		return false
	}
	return !stat.IsDir()
}

func EnsureDirectoryExists(directory string) error {
	_, err := EnsureDirectory(directory)
	return err
}

func EnsureEmptyDirectory(directory string) error {
	fullpath, err := EnsureDirectory(directory)
	handle, err := os.Open(fullpath)
	if err != nil {
		return err
	}
	entries, err := handle.Readdir(-1)
	if len(entries) > 0 {
		return errors.New(fmt.Sprintf("Directory %s is not empty!", fullpath))
	}
	return nil
}
