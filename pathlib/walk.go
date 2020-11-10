package pathlib

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Ignore func(os.FileInfo) bool
type Report func(string, string, os.FileInfo)

type IgnoreOlder time.Time

func (it IgnoreOlder) Ignore(candidate os.FileInfo) bool {
	return candidate.ModTime().Before(time.Time(it))
}

type IgnoreNewer time.Time

func (it IgnoreNewer) Ignore(candidate os.FileInfo) bool {
	return candidate.ModTime().After(time.Time(it))
}

func IgnoreNothing(_ os.FileInfo) bool {
	return false
}

func IgnoreDirectories(target os.FileInfo) bool {
	return target.IsDir()
}

func NoReporting(string, string, os.FileInfo) {
}

func sorted(files []os.FileInfo) {
	sort.SliceStable(files, func(left, right int) bool {
		return files[left].Name() < files[right].Name()
	})
}

type composite []Ignore

func (it composite) Ignore(file os.FileInfo) bool {
	for _, ignore := range it {
		if ignore(file) {
			return true
		}
	}
	return false
}

type exactIgnore string

func (it exactIgnore) Ignore(file os.FileInfo) bool {
	return file.Name() == string(it)
}

type globIgnore string

func (it globIgnore) Ignore(file os.FileInfo) bool {
	name := file.Name()
	result, err := filepath.Match(string(it), name)
	if err == nil && result {
		return true
	}
	if file.IsDir() {
		result, err = filepath.Match(string(it), name+"/")
		return err == nil && result
	}
	return false
}

func CompositeIgnore(ignores ...Ignore) Ignore {
	return composite(ignores).Ignore
}

func IgnorePattern(text string) Ignore {
	return CompositeIgnore(exactIgnore(text).Ignore, globIgnore(text).Ignore)
}

func LoadIgnoreFile(filename string) (Ignore, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	result := make([]Ignore, 0, 10)
	for _, line := range strings.SplitAfter(string(content), "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		result = append(result, IgnorePattern(line))
	}
	return CompositeIgnore(result...), nil
}

func LoadIgnoreFiles(filenames []string) (Ignore, error) {
	result := make([]Ignore, 0, len(filenames))
	for _, filename := range filenames {
		ignore, err := LoadIgnoreFile(filename)
		if err != nil {
			return nil, err
		}
		result = append(result, ignore)
	}
	return CompositeIgnore(result...), nil
}

func folderEntries(directory string) ([]os.FileInfo, error) {
	handle, err := os.Open(directory)
	if err != nil {
		return nil, err
	}
	defer handle.Close()
	entries, err := handle.Readdir(-1)
	if err != nil {
		return nil, err
	}
	return entries, nil
}

func recursiveWalk(directory, prefix string, ignore Ignore, report Report) error {
	entries, err := folderEntries(directory)
	if err != nil {
		return err
	}
	sorted(entries)
	for _, entry := range entries {
		if ignore(entry) {
			continue
		}
		nextPrefix := filepath.Join(prefix, entry.Name())
		entryPath := filepath.Join(directory, entry.Name())
		if entry.IsDir() {
			recursiveWalk(entryPath, nextPrefix, ignore, report)
		} else {
			report(entryPath, nextPrefix, entry)
		}
	}
	return nil
}

func Walk(directory string, ignore Ignore, report Report) error {
	fullpath, err := filepath.Abs(directory)
	if err != nil {
		return err
	}
	return recursiveWalk(fullpath, ".", ignore, report)
}
