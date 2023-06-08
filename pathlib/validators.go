package pathlib

import (
	"fmt"
	"os"

	"github.com/robocorp/rcc/pretty"
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
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(fullpath)
	if err != nil {
		return err
	}
	if len(entries) > 0 {
		return fmt.Errorf("Directory %s is not empty!", fullpath)
	}
	return nil
}

func NoteDirectoryContent(context, directory string) {
	if !IsDir(directory) {
		return
	}
	fullpath, err := Abs(directory)
	if err != nil {
		return
	}
	entries, err := os.ReadDir(fullpath)
	if err != nil {
		return
	}
	for _, entry := range entries {
		pretty.Note("%s %q already has %q in it.", context, fullpath, entry.Name())
	}
}
