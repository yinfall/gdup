package gdup

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

// VersionInfo holds parsed version details for sorting.
type VersionInfo struct {
	Major      int
	Minor      int
	Patch      int
	StatusRank int
	StatusNum  int
	Original   string
}

var statusOrder = map[string]int{
	"stable": 5,
	"rc":     4,
	"beta":   3,
	"alpha":  2,
	"dev":    1,
}

var versionRegex = regexp.MustCompile(`(?i)godot_v([0-9]+)\.([0-9]+)(?:\.([0-9]+))?(?:-([a-zA-Z0-9.]+))?`)
var versionStrRegex = regexp.MustCompile(`(?i)godot_v([0-9]+\.[0-9]+(?:\.[0-9]+)?(?:-[a-zA-Z0-9.]+)?)(?:_|\.exe)`)
var bareVersionRegex = regexp.MustCompile(`^\d+\.\d+(?:\.\d+)*$`)

// ParseVersionFromFilename extracts version details from a Godot filename.
func ParseVersionFromFilename(filename string) VersionInfo {
	match := versionRegex.FindStringSubmatch(filename)
	if match == nil {
		return VersionInfo{}
	}

	major := atoi(match[1])
	minor := atoi(match[2])
	patch := atoi(match[3])
	status := "stable"
	if match[4] != "" {
		status = match[4]
	}

	statusType := lettersOnly(strings.ToLower(status))
	statusNum := trailingDigits(status)

	rank, ok := statusOrder[statusType]
	if !ok {
		rank = 0
	}

	vStr := ""
	vMatch := versionStrRegex.FindStringSubmatch(filename)
	if vMatch != nil {
		vStr = vMatch[1]
	}

	return VersionInfo{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		StatusRank: rank,
		StatusNum:  statusNum,
		Original:   vStr,
	}
}

// NormalizeVersion adds -stable suffix to bare version numbers.
func NormalizeVersion(version string) string {
	tag := strings.TrimPrefix(version, "v")
	if bareVersionRegex.MatchString(tag) {
		tag = tag + "-stable"
	}
	return tag
}

// NormalizeVersionTag trims 'v' prefix and lowercases.
func NormalizeVersionTag(version string) string {
	v := strings.TrimPrefix(version, "v")
	return strings.ToLower(v)
}

// VersionMatches checks if a version string matches a target (case-insensitive, ignoring -stable).
func VersionMatches(v, target string) bool {
	v = strings.ToLower(v)
	target = strings.ToLower(target)
	if v == target {
		return true
	}
	return strings.ReplaceAll(v, "-stable", "") == strings.ReplaceAll(target, "-stable", "")
}

// CompareVersionInfo compares two VersionInfo for sorting.
func CompareVersionInfo(a, b VersionInfo) int {
	if a.Major != b.Major {
		return a.Major - b.Major
	}
	if a.Minor != b.Minor {
		return a.Minor - b.Minor
	}
	if a.Patch != b.Patch {
		return a.Patch - b.Patch
	}
	if a.StatusRank != b.StatusRank {
		return a.StatusRank - b.StatusRank
	}
	return a.StatusNum - b.StatusNum
}

// LatestByVersion returns the latest file from a list by version sorting.
func LatestByVersion(files []string) string {
	sort.Slice(files, func(i, j int) bool {
		vi := ParseVersionFromFilename(files[i])
		vj := ParseVersionFromFilename(files[j])
		return CompareVersionInfo(vi, vj) < 0
	})
	if len(files) > 0 {
		return files[len(files)-1]
	}
	return ""
}

// GetGodotDir returns the godot-versions directory path.
func GetGodotDir() string {
	if cachePath := os.Getenv("GDUP_CACHE_PATH"); cachePath != "" {
		return filepath.Join(cachePath, "versions")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to executable directory if home is unavailable
		exe, _ := os.Executable()
		return filepath.Join(filepath.Dir(exe), "versions")
	}
	return filepath.Join(home, ".gdup", "versions")
}

// PlatformSuffix returns the expected suffix for the current platform.
func PlatformSuffix() string {
	switch runtime.GOOS + "/" + runtime.GOARCH {
	case "windows/amd64":
		return "_win64"
	case "windows/386":
		return "_win32"
	case "darwin/amd64":
		return "_macos"
	case "darwin/arm64":
		return "_macos"
	case "linux/amd64":
		return "_linux"
	default:
		return "_win64"
	}
}

// ExeExtension returns the executable extension for the current OS.
func ExeExtension() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			break
		}
	}
	return n
}

func lettersOnly(s string) string {
	var b strings.Builder
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			b.WriteRune(c)
		}
	}
	return b.String()
}

func trailingDigits(s string) int {
	re := regexp.MustCompile(`\d+$`)
	m := re.FindString(s)
	if m == "" {
		return 0
	}
	return atoi(m)
}
