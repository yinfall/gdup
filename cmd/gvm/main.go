package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/user/gvm/internal/gvmcore"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	command := strings.ToLower(os.Args[1])

	switch command {
	case "--help", "-h", "help":
		printHelp()
	case "list", "ls":
		cmdList()
	case "releases", "release":
		showAll := false
		if len(os.Args) > 2 && (os.Args[2] == "-a" || os.Args[2] == "--all") {
			showAll = true
		}
		cmdReleases(showAll)
	case "install":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: Please specify the version to install. E.g.: gvm install 4.6.3")
			os.Exit(1)
		}
		cmdInstall(os.Args[2])
	case "use":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: Please specify the version to use. E.g.: gvm use 4.6.3")
			os.Exit(1)
		}
		cmdUse(os.Args[2])
	case "uninstall", "remove":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: Please specify the version to uninstall. E.g.: gvm uninstall 4.6.3")
			os.Exit(1)
		}
		force := len(os.Args) > 3 && (os.Args[3] == "-y" || os.Args[3] == "--yes")
		cmdUninstall(os.Args[2], force)
	case "--version", "-v":
		fmt.Printf("gvm %s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'.\n", command)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	godotDir := gvmcore.GetGodotDir()
	fmt.Printf(`Godot Version Manager (GVM) %s

Usage:
  gvm list | ls                     Show locally installed Godot versions in %s
  gvm releases [-a]                 List available releases from GitHub (use -a/--all for pre-releases)
  gvm install <version>             Download and install a specific Godot version (e.g. 4.6.3)
  gvm use <version>                 Set the active Godot version for the current directory
  gvm uninstall <version> [-y]      Uninstall a locally installed version (-y to skip confirmation)
  gvm --version                     Show GVM version

Example:
  gvm install 4.6.3
  gvm use 4.6.3
  gvm uninstall 4.6.3
`, version, godotDir)
}

func cmdList() {
	godotDir := gvmcore.GetGodotDir()
	if _, err := os.Stat(godotDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Godot installation directory '%s' does not exist.\n", godotDir)
		os.Exit(1)
	}

	installed, err := gvmcore.GetInstalledVersions(godotDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(installed) == 0 {
		fmt.Printf("No Godot versions installed in %s.\n", godotDir)
		return
	}

	activeVersion := gvmcore.GetActiveVersion()

	fmt.Printf("Installed versions in %s:\n", godotDir)
	for _, iv := range installed {
		isActive := gvmcore.VersionMatches(iv.Version, activeVersion)

		activeLabel := ""
		prefix := "    "
		if isActive {
			activeLabel = " (active)"
			prefix = "  * "
		}

		hasConsole := false
		hasGUI := false
		for _, f := range iv.Files {
			if strings.Contains(strings.ToLower(f), "_console") {
				hasConsole = true
			} else {
				hasGUI = true
			}
		}

		var types []string
		if hasGUI {
			types = append(types, "GUI")
		}
		if hasConsole {
			types = append(types, "Console")
		}
		typesStr := ""
		if len(types) > 0 {
			typesStr = " [" + strings.Join(types, ", ") + "]"
		}

		fmt.Printf("%s%s%s%s\n", prefix, iv.Version, typesStr, activeLabel)
	}
}

// GitHub API types
type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

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
	req.Header.Set("User-Agent", "GVM/"+version)
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

func cmdReleases(showAll bool) {
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

func cmdInstall(version string) {
	godotDir := gvmcore.GetGodotDir()
	os.MkdirAll(godotDir, 0755)

	tag := gvmcore.NormalizeVersion(version)

	// Check if already installed
	installed, _ := gvmcore.GetInstalledVersions(godotDir)
	for _, iv := range installed {
		if gvmcore.VersionMatches(iv.Version, tag) {
			fmt.Printf("Version '%s' is already installed.\n", tag)
			return
		}
	}

	fmt.Printf("Searching for release '%s' on GitHub...\n", tag)
	repo := getRepoForTag(tag)
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", repo, tag)

	var release githubRelease
	if err := fetchJSON(url, &release); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Release '%s' not found or GitHub API error: %v\n", tag, err)
		fmt.Fprintln(os.Stderr, "Please verify the version number. Use 'gvm releases -a' to list all available versions.")
		os.Exit(1)
	}

	// Find suitable asset for current platform
	platformSuffix := gvmcore.PlatformSuffix()
	exeExt := gvmcore.ExeExtension()
	var downloadURL, assetName string

	for _, a := range release.Assets {
		name := strings.ToLower(a.Name)
		// Match platform-specific zip, exclude mono and debug
		if strings.Contains(name, platformSuffix) &&
			strings.HasSuffix(name, ".zip") &&
			!strings.Contains(name, "mono") &&
			!strings.Contains(name, "debug") {
			// Prefer .exe.zip for Windows
			if runtime.GOOS == "windows" {
				if strings.HasSuffix(name, ".exe.zip") {
					downloadURL = a.BrowserDownloadURL
					assetName = a.Name
					break
				}
			} else {
				downloadURL = a.BrowserDownloadURL
				assetName = a.Name
				break
			}
		}
	}

	// Fallback: for Windows, also accept non-.exe.zip
	if downloadURL == "" && runtime.GOOS == "windows" {
		for _, a := range release.Assets {
			name := strings.ToLower(a.Name)
			if strings.Contains(name, platformSuffix) &&
				strings.HasSuffix(name, ".zip") &&
				!strings.Contains(name, "mono") &&
				!strings.Contains(name, "debug") {
				downloadURL = a.BrowserDownloadURL
				assetName = a.Name
				break
			}
		}
	}

	if downloadURL == "" {
		fmt.Fprintf(os.Stderr, "Error: Could not find a suitable %s build in release '%s'.\n", platformSuffix, tag)
		fmt.Fprintln(os.Stderr, "Available assets for this release:")
		for _, a := range release.Assets {
			fmt.Fprintf(os.Stderr, "  - %s\n", a.Name)
		}
		os.Exit(1)
	}

	// Download to temp file
	tempDir := os.TempDir()
	tempZipPath := filepath.Join(tempDir, fmt.Sprintf("gvm_download_%s.zip", tag))

	fmt.Printf("Found asset: %s\n", assetName)
	if !downloadWithProgress(downloadURL, tempZipPath) {
		os.Exit(1)
	}

	// Extract and install
	if !extractZipAndClean(tempZipPath, godotDir) {
		os.Exit(1)
	}

	// Clean up zip
	os.Remove(tempZipPath)

	fmt.Printf("Successfully installed Godot %s!\n", tag)
	_ = exeExt // suppress unused warning
}

func downloadWithProgress(url, destPath string) bool {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		return false
	}
	req.Header.Set("User-Agent", "GVM/"+version)

	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error downloading file: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Fprintf(os.Stderr, "Error: server returned status %d\n", resp.StatusCode)
		return false
	}

	totalSize := resp.ContentLength
	f, err := os.Create(destPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
		return false
	}
	defer f.Close()

	fmt.Println("Downloading...")
	start := time.Now()
	buf := make([]byte, 64*1024)
	var downloaded int64

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				fmt.Fprintf(os.Stderr, "Error writing file: %v\n", werr)
				os.Remove(destPath)
				return false
			}
			downloaded += int64(n)
		}

		if totalSize > 0 {
			percent := float64(downloaded) / float64(totalSize)
			barLength := 30
			filled := int(float64(barLength) * percent)
			bar := strings.Repeat("█", filled) + strings.Repeat("-", barLength-filled)
			elapsed := time.Since(start).Seconds()
			speed := float64(0)
			if elapsed > 0 {
				speed = (float64(downloaded) / 1024 / 1024) / elapsed
			}
			fmt.Printf("\r|%s| %.1f%% (%.1f/%.1f MB) - %.1f MB/s",
				bar, percent*100,
				float64(downloaded)/1024/1024, float64(totalSize)/1024/1024,
				speed)
		} else {
			fmt.Printf("\rDownloaded %.1f MB", float64(downloaded)/1024/1024)
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nError downloading: %v\n", err)
			os.Remove(destPath)
			return false
		}
	}
	fmt.Println("\nDownload complete.")
	return true
}

func extractZipAndClean(zipPath, extractDir string) bool {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening zip: %v\n", err)
		return false
	}
	defer r.Close()

	fmt.Println("Extracting package...")

	for _, f := range r.File {
		// Skip directories
		if f.FileInfo().IsDir() {
			continue
		}

		// Handle paths: strip leading directory if all files share one prefix
		name := f.Name
		// Clean the path to prevent zip slip
		name = filepath.Clean(name)
		if strings.HasPrefix(name, "..") {
			continue
		}

		// If the file is inside a single directory, strip it
		parts := strings.SplitN(name, string(filepath.Separator), 2)
		if len(parts) == 2 {
			name = parts[1]
		}

		destPath := filepath.Join(extractDir, name)
		os.MkdirAll(filepath.Dir(destPath), 0755)

		rc, err := f.Open()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file in zip: %v\n", err)
			return false
		}

		outFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
			return false
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			return false
		}
	}

	fmt.Println("Installation complete.")
	return true
}

func cmdUse(version string) {
	godotDir := gvmcore.GetGodotDir()
	if _, err := os.Stat(godotDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Godot installation directory '%s' does not exist.\n", godotDir)
		os.Exit(1)
	}

	tag := gvmcore.NormalizeVersion(version)

	installed, err := gvmcore.GetInstalledVersions(godotDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var matched *gvmcore.InstalledVersion
	for i := range installed {
		if gvmcore.VersionMatches(installed[i].Version, tag) {
			matched = &installed[i]
			break
		}
	}

	if matched == nil {
		fmt.Fprintf(os.Stderr, "Error: Version '%s' is not installed locally in %s.\n", version, godotDir)
		fmt.Fprintln(os.Stderr, "Please install it first using: gvm install <version>")
		os.Exit(1)
	}

	// Write .gvm in current directory
	cwd, _ := os.Getwd()
	configPath := filepath.Join(cwd, ".gvm")
	configData := map[string]string{"version": matched.Version}

	data, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to serialize config: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to write configuration file '%s': %v\n", configPath, err)
		os.Exit(1)
	}

	fmt.Printf("Success: Active Godot version set to '%s' in this directory (%s).\n", matched.Version, configPath)
}

func cmdUninstall(version string, force bool) {
	godotDir := gvmcore.GetGodotDir()
	if _, err := os.Stat(godotDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Godot installation directory '%s' does not exist.\n", godotDir)
		os.Exit(1)
	}

	tag := gvmcore.NormalizeVersion(version)

	installed, err := gvmcore.GetInstalledVersions(godotDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var matched *gvmcore.InstalledVersion
	for i := range installed {
		if gvmcore.VersionMatches(installed[i].Version, tag) {
			matched = &installed[i]
			break
		}
	}

	if matched == nil {
		fmt.Fprintf(os.Stderr, "Error: Version '%s' is not installed.\n", version)
		os.Exit(1)
	}

	// Warn if active
	activeVersion := gvmcore.GetActiveVersion()
	if gvmcore.VersionMatches(matched.Version, activeVersion) {
		configPath := gvmcore.FindGvmConfig()
		fmt.Printf("Warning: '%s' is currently the active version in '%s'.\n", matched.Version, configPath)
	}

	fmt.Println("The following files will be removed:")
	for _, f := range matched.Files {
		fmt.Printf("  - %s\n", f)
	}

	if !force {
		fmt.Printf("\nUninstall Godot %s? [y/N] ", matched.Version)
		var answer string
		fmt.Scanln(&answer)
		if strings.ToLower(answer) != "y" && strings.ToLower(answer) != "yes" {
			fmt.Println("Aborted.")
			os.Exit(0)
		}
	}

	hasError := false
	for _, f := range matched.Files {
		p := filepath.Join(godotDir, f)
		if err := os.Remove(p); err != nil {
			fmt.Fprintf(os.Stderr, "  - %s: %v\n", f, err)
			hasError = true
		}
	}

	if hasError {
		fmt.Fprintln(os.Stderr, "Some files could not be deleted.")
		os.Exit(1)
	}

	fmt.Printf("Successfully uninstalled Godot %s!\n", matched.Version)
}
