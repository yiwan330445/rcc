package pathlib

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type PathParts []string

func TargetPath() PathParts {
	return filepath.SplitList(os.Getenv("PATH"))
}

func PathFrom(parts ...string) PathParts {
	if parts == nil {
		return PathParts{}
	}
	return parts
}

func (it PathParts) AsEnvironmental(name string) string {
	return fmt.Sprintf("%s=%s", name, strings.Join(it.Absolute(), string(filepath.ListSeparator)))
}

func (it PathParts) Remove(patterns []string) PathParts {
	result := make([]string, 0, len(it))
	for _, path := range it {
		reference := strings.ToLower(path)
		skip := false
		for _, ignore := range patterns {
			if strings.Index(reference, ignore) > -1 {
				skip = true
				break
			}
		}
		if !skip {
			result = append(result, path)
		}
	}
	return result
}

func (it PathParts) Append(parts ...string) PathParts {
	result := make([]string, 0, len(it)+len(parts))
	result = append(result, it...)
	result = append(result, parts...)
	return result
}

func (it PathParts) Prepend(parts ...string) PathParts {
	result := make([]string, 0, len(it)+len(parts))
	result = append(result, parts...)
	result = append(result, it...)
	return result
}

func (it PathParts) Absolute() PathParts {
	result := make([]string, 0, len(it))
	for _, item := range it {
		full, err := filepath.Abs(item)
		if err == nil {
			result = append(result, full)
		}
	}
	return result
}

func whichVariation(fullPath string, fileExtensions []string) (string, bool) {
	for _, extension := range fileExtensions {
		variation := fullPath + extension
		stat, err := os.Stat(variation)
		if err == nil && !stat.IsDir() {
			return variation, true
		}
	}
	return "", false
}

func (it PathParts) Which(application string, extensions []string) (string, bool) {
	if filepath.IsAbs(application) && IsFile(application) {
		return application, true
	}
	for _, directory := range it {
		stat, err := os.Stat(directory)
		if err != nil || !stat.IsDir() {
			continue
		}
		found, ok := whichVariation(filepath.Join(directory, application), extensions)
		if ok {
			return found, ok
		}
	}
	return "", false
}
