package gdup

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// parseAssetFilename extracts release type and phase from an asset filename.
// Examples:
//   "Godot_v4.8-dev2_linux.x86_64.zip" -> ("dev", "dev2")
//   "Godot_v4.7.2-rc1_windows_x86_64.zip" -> ("rc", "rc1")
//   "Godot_v4.7.1-stable_linux.x86_64.zip" -> ("stable", "stable")
func parseAssetFilename(filename string) (string, string) {
	// Look for known release types in the filename
	for _, typ := range []string{"dev", "rc", "beta", "alpha", "stable"} {
		idx := strings.Index(strings.ToLower(filename), typ)
		if idx != -1 {
			// Extract from typ start to end of consecutive digits
			end := idx + len(typ)
			for end < len(filename) && filename[end] >= '0' && filename[end] <= '9' {
				end++
			}
			phase := filename[idx:end]
			return typ, phase
		}
	}
	return "stable", "stable"
}

// parseVersionTag extracts version number and release type from a tag.
// Examples: "v4.6.3" -> ("4.6.3", "stable")
//           "4.6.3-stable" -> ("4.6.3", "stable")
//           "4.6.3-rc1" -> ("4.6.3", "rc")
//           "4.6.3-beta2" -> ("4.6.3", "beta")
func parseVersionTag(tag string) (string, string) {
	// Remove leading 'v'
	tag = strings.TrimPrefix(tag, "v")

	// Check for known release types (including stable)
	for _, typ := range []string{"stable", "rc", "beta", "dev", "alpha"} {
		if idx := strings.Index(strings.ToLower(tag), typ); idx != -1 {
			version := tag[:idx]
			// Remove trailing dash if present
			version = strings.TrimRight(version, "-")
			return version, typ
		}
	}

	// Default to stable if no type found
	return tag, "stable"
}

// colorize returns the ANSI color code for a release type.
func colorize(releaseType string) string {
	switch releaseType {
	case "stable":
		return "\033[32m" // Green
	case "rc":
		return "\033[33m" // Yellow
	case "beta":
		return "\033[34m" // Blue
	case "dev":
		return "\033[31m" // Red
	case "alpha":
		return "\033[35m" // Magenta
	default:
		return "\033[0m" // Reset
	}
}

const (
	resetColor = "\033[0m"
	border     = "+----------------+----------------+--------------------+"
	separator   = "|" + "----------------" + "|" + "----------------" + "|" + "--------------------" + "|"
)

// githubRelease represents a GitHub release.
type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

// githubAsset represents a release asset on GitHub.
type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func getRepoForTag(tag string) string {
	tagLower := strings.ToLower(tag)
	for _, pre := range []string{"rc", "beta", "dev", "alpha"} {
		if strings.Contains(tagLower, pre) {
			return "godotengine/godot-builds"
		}
	}
	return "godotengine/godot"
}

func fetchJSON(url string, v interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "GDUP/"+GDUPVersion)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

// CmdReleases lists available releases from GitHub.
func CmdReleases(showAll bool) {
	repo := "godotengine/godot"
	label := "stable"
	if showAll {
		repo = "godotengine/godot-builds"
		label = "all"
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/releases?per_page=100", repo)
	fmt.Printf("Fetching %s releases from GitHub (%s)...\n", label, repo)

	var releases []githubRelease
	if err := fetchJSON(url, &releases); err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching releases: %v\n", err)
		os.Exit(1)
	}

	// Filter releases based on showAll flag
	var filtered []githubRelease
	for _, r := range releases {
		if showAll || strings.Contains(strings.ToLower(r.TagName), "stable") {
			filtered = append(filtered, r)
		}
	}

	if len(filtered) == 0 {
		fmt.Printf("\nNo %s releases found.\n", label)
		return
	}

	// Print table header
	fmt.Printf("\nAvailable %s releases:\n\n", label)
	fmt.Println(border)
	fmt.Printf("| %-14s | %-14s | %-18s |\n", "VERSION", "TYPE", "COMMAND")
	fmt.Println(separator)

	// Print each release
	for _, r := range filtered {
		// Get version from tag
		version, _ := parseVersionTag(r.TagName)

		// Try to get phase from asset filename first
		releaseType, phase := "stable", "stable"
		if len(r.Assets) > 0 {
			// Find an asset for the current platform (linux)
			for _, asset := range r.Assets {
				if strings.Contains(strings.ToLower(asset.Name), "linux") {
					releaseType, phase = parseAssetFilename(asset.Name)
					break
				}
			}
			// If no linux asset, use the first asset
			if releaseType == "stable" && phase == "stable" && len(r.Assets) > 0 {
				releaseType, phase = parseAssetFilename(r.Assets[0].Name)
			}
		} else {
			// Fallback to tag parsing
			_, releaseType = parseVersionTag(r.TagName)
			phase = releaseType
		}

		// If phase is empty, fallback to releaseType
		if phase == "" {
			phase = releaseType
		}

		// Generate install command
		var installCmd string
		if releaseType == "stable" {
			installCmd = fmt.Sprintf("install %s", version)
		} else {
			installCmd = fmt.Sprintf("install %s-%s", version, strings.ToLower(phase))
		}

		color := colorize(releaseType)
		fmt.Printf("| %s%-14s%s | %s%-14s%s | %-18s |\n", color, version, resetColor, color, strings.ToUpper(phase), resetColor, installCmd)
	}

	// Print table footer
	fmt.Println(border)
}
