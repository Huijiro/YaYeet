package patchnotes

import (
	"embed"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

//go:embed *.md
var notes embed.FS

func ForVersion(version string) (string, bool) {
	if path.Base(version) != version {
		return "", false
	}
	content, err := notes.ReadFile(version + ".md")
	if err != nil {
		return "", false
	}
	return string(content), true
}

func PreviousForVersion(version string) (string, string, bool) {
	if path.Base(version) != version {
		return "", "", false
	}
	selected := versionKey(version)
	entries, err := notes.ReadDir(".")
	if err != nil {
		return "", "", false
	}
	candidates := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := strings.TrimSuffix(entry.Name(), ".md")
		if entry.IsDir() || name == version || versionKey(name).compare(selected) >= 0 {
			continue
		}
		candidates = append(candidates, name)
	}
	if len(candidates) == 0 {
		return "", "", false
	}
	sort.Slice(candidates, func(i, j int) bool {
		return versionKey(candidates[i]).compare(versionKey(candidates[j])) > 0
	})
	previous := candidates[0]
	content, err := notes.ReadFile(previous + ".md")
	if err != nil {
		return "", "", false
	}
	return previous, string(content), true
}

type versionParts struct {
	numbers []int
	suffix  string
}

var versionPattern = regexp.MustCompile(`^(\\d+(?:\\.\\d+)*)(.*)$`)

func versionKey(version string) versionParts {
	match := versionPattern.FindStringSubmatch(version)
	if match == nil {
		return versionParts{suffix: version}
	}
	parts := strings.Split(match[1], ".")
	numbers := make([]int, len(parts))
	for index, part := range parts {
		numbers[index], _ = strconv.Atoi(part)
	}
	return versionParts{numbers: numbers, suffix: match[2]}
}

func (version versionParts) compare(other versionParts) int {
	for index := 0; index < len(version.numbers) && index < len(other.numbers); index++ {
		if version.numbers[index] != other.numbers[index] {
			if version.numbers[index] < other.numbers[index] {
				return -1
			}
			return 1
		}
	}
	if len(version.numbers) != len(other.numbers) {
		if len(version.numbers) < len(other.numbers) {
			return -1
		}
		return 1
	}
	return strings.Compare(version.suffix, other.suffix)
}
