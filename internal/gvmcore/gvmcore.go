package gvmcore

import (
	"encoding/json"
	"fmt"
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
	Original   string // the version string as found in filename
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

	// Extract letters for status ranking
	statusType := lettersOnly(strings.ToLower(status))
	statusNum := trailingDigits(status)

	rank, ok := statusOrder[statusType]
	if !ok {
		rank = 0
	}

	// Extract the version string
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

// VersionMatches checks if a version string matches a target (case-insensitive, ignoring -stable).
func VersionMatches(v, target string) bool {
	v = strings.ToLower(v)
	target = strings.ToLower(target)
	if v == target {
		return true
	}
	return strings.ReplaceAll(v, "-stable", "") == strings.ReplaceAll(target, "-stable", "")
}

// GvmConfig represents the .gvm configuration file.
type GvmConfig struct {
	Version string `json:"version"`
}

// FindGvmConfig searches for .gvm file from cwd upwards, then in home directory.
func FindGvmConfig() string {
	cwd, err := os.Getwd()
	if err == nil {
		dir := cwd
		for {
			p := filepath.Join(dir, ".gvm")
			if info, err := os.Stat(p); err == nil && !info.IsDir() {
				return p
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	home, err := os.UserHomeDir()
	if err == nil {
		p := filepath.Join(home, ".gvm")
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p
		}
	}

	return ""
}

// ReadGvmConfig reads and parses a .gvm config file.
func ReadGvmConfig(path string) (*GvmConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg GvmConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// GetActiveVersion returns the active version from .gvm config, or empty string.
func GetActiveVersion() string {
	configPath := FindGvmConfig()
	if configPath == "" {
		return ""
	}
	cfg, err := ReadGvmConfig(configPath)
	if err != nil {
		return ""
	}
	v := cfg.Version
	v = strings.TrimPrefix(v, "v")
	return strings.ToLower(v)
}

// InstalledVersion holds info about an installed version.
type InstalledVersion struct {
	Version string   // e.g. "4.6.3-stable"
	Files   []string // filenames
}

// GetInstalledVersions scans the godot-versions directory.
func GetInstalledVersions(godotDir string) ([]InstalledVersion, error) {
	entries, err := os.ReadDir(godotDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory '%s': %w", godotDir, err)
	}

	groups := map[string][]string{}
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(strings.ToLower(name), ".exe") {
			vInfo := ParseVersionFromFilename(name)
			if vInfo.Original != "" {
				groups[vInfo.Original] = append(groups[vInfo.Original], name)
			}
		}
	}

	var result []InstalledVersion
	for v, files := range groups {
		result = append(result, InstalledVersion{Version: v, Files: files})
	}

	// Sort descending
	sort.Slice(result, func(i, j int) bool {
		vi := ParseVersionFromFilename(result[i].Files[0])
		vj := ParseVersionFromFilename(result[j].Files[0])
		return compareVersionInfo(vi, vj) > 0
	})

	return result, nil
}

func compareVersionInfo(a, b VersionInfo) int {
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

// GetGodotDir returns the godot-versions directory path.
func GetGodotDir() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Join(filepath.Dir(exe), "godot-versions")
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
		return "_win64" // fallback
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
