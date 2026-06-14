package gvm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
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
	req.Header.Set("User-Agent", "GVM/"+GVMVersion)
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

	fmt.Printf("\nAvailable %s releases:\n", label)
	for _, r := range releases {
		tag := strings.TrimPrefix(r.TagName, "v")
		if showAll || strings.Contains(strings.ToLower(r.TagName), "stable") {
			fmt.Printf("  - %s\n", tag)
		}
	}
}
