package conda

import (
	"errors"
	"io/ioutil"
	"regexp"
	"sort"
	"strings"

	"github.com/glaslos/tlsh"
)

var (
	linebreaks = regexp.MustCompile("\r?\n")
)

func UnifyLine(value string) string {
	return strings.Trim(value, " \t\r\n")
}

func SplitLines(value string) []string {
	return linebreaks.Split(value, -1)
}

func ReadConfig(filename string) (string, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func LocalitySensitiveHash(parts []string) (string, error) {
	content := "==================================================\n"
	content += strings.Join(parts, "\n")
	result, err := tlsh.HashBytes([]byte(content))
	return result.String(), err
}

func AsUnifiedLines(value string) []string {
	parts := SplitLines(value)
	limit := len(parts)
	seen := make(map[string]bool, limit)
	result := make([]string, 0, limit)
	for _, part := range parts {
		unified := UnifyLine(part)
		if seen[unified] {
			continue
		}
		seen[unified] = true
		if len(unified) > 0 {
			result = append(result, unified)
		}
	}
	sort.Strings(result)
	return result
}

func HashConfig(filename string) (string, error) {
	content, err := ReadConfig(filename)
	if err != nil {
		return "", err
	}
	hash, err := LocalitySensitiveHash(AsUnifiedLines(content))
	return hash, err
}

func Distance(left, right string) (int, error) {
	if len(left) != 70 || len(left) != len(right) {
		return 999999, errors.New("Incorrect length of TLSH hashes.")
	}
	leftish, err := tlsh.ParseStringToTlsh(left)
	if err != nil {
		return 0, err
	}
	rightish, err := tlsh.ParseStringToTlsh(right)
	if err != nil {
		return 0, err
	}
	return leftish.Diff(rightish), nil
}
