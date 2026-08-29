package gdup

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// CmdInstall downloads and installs a specific Godot version.
func CmdInstall(version string) {
	godotDir := GetGodotDir()
	os.MkdirAll(godotDir, 0755)

	tag := NormalizeVersion(version)

	// Check if already installed
	installed, _ := GetInstalledVersions(godotDir)
	for _, iv := range installed {
		if VersionMatches(iv.Version, tag) {
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
		fmt.Fprintln(os.Stderr, "Please verify the version number. Use 'godot releases -a' to list all available versions.")
		os.Exit(1)
	}

	// Find suitable asset for current platform
	platformSuffix := PlatformSuffix()
	var downloadURL, assetName string

	for _, a := range release.Assets {
		name := strings.ToLower(a.Name)
		if strings.Contains(name, platformSuffix) &&
			strings.HasSuffix(name, ".zip") &&
			!strings.Contains(name, "mono") &&
			!strings.Contains(name, "debug") {
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
	tempZipPath := filepath.Join(os.TempDir(), fmt.Sprintf("gdup_download_%s.zip", tag))

	fmt.Printf("Found asset: %s\n", assetName)
	if !downloadWithProgress(downloadURL, tempZipPath) {
		os.Exit(1)
	}

	// Extract and install
	if !extractZipAndClean(tempZipPath, godotDir) {
		os.Exit(1)
	}

	os.Remove(tempZipPath)
	fmt.Printf("Successfully installed Godot %s!\n", tag)
}

func downloadWithProgress(url, destPath string) bool {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		return false
	}
	req.Header.Set("User-Agent", "GDUP/"+GDUPVersion)

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
		if f.FileInfo().IsDir() {
			continue
		}

		name := filepath.Clean(f.Name)
		if strings.HasPrefix(name, "..") {
			continue
		}

		// Strip leading directory
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
