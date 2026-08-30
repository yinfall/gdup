package gdup

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// InstalledVersion holds info about an installed version.
type InstalledVersion struct {
	Version string
	Files   []string
}

// GetInstalledVersions scans the godot-versions directory.
func GetInstalledVersions(godotDir string) ([]InstalledVersion, error) {
	entries, err := os.ReadDir(godotDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory '%s': %w", godotDir, err)
	}

	var result []InstalledVersion
	for _, e := range entries {
		if !e.IsDir() {
			continue // skip files (should be migrated or ignored)
		}
		
		versionName := e.Name()
		versionDir := filepath.Join(godotDir, versionName)
		
		var files []string
		
		filepath.WalkDir(versionDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			
			if d.IsDir() {
				return nil
			}

			name := strings.ToLower(d.Name())
			isExecutable := false

			if runtime.GOOS == "windows" {
				isExecutable = strings.HasSuffix(name, ".exe")
			} else {
				info, err := d.Info()
				if err == nil && (info.Mode()&0111 != 0) {
					isExecutable = true
				}
				if strings.Contains(name, "godot") && !strings.HasSuffix(name, ".pck") && !strings.HasSuffix(name, ".txt") && !strings.HasSuffix(name, ".md") {
					isExecutable = true
				}
			}

			if isExecutable {
				relPath, _ := filepath.Rel(versionDir, path)
				files = append(files, relPath)
			}
			return nil
		})
		
		if len(files) > 0 {
			result = append(result, InstalledVersion{Version: versionName, Files: files})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		vi := ParseVersion(result[i].Version)
		vj := ParseVersion(result[j].Version)
		comp := CompareVersionInfo(vi, vj)
		if comp == 0 {
			// fallback to string compare
			return result[i].Version < result[j].Version
		}
		return comp < 0
	})

	return result, nil
}

func parseInstalledVersion(raw string) (string, string, string) {
	releaseType, phase := parseAssetFilename(raw)
	isMono := strings.Contains(strings.ToLower(raw), "mono")

	vi := ParseVersionFromFilename(raw)
	var verNum string
	if vi.Major > 0 {
		if vi.Patch > 0 {
			verNum = fmt.Sprintf("%d.%d.%d", vi.Major, vi.Minor, vi.Patch)
		} else {
			verNum = fmt.Sprintf("%d.%d", vi.Major, vi.Minor)
		}
	} else {
		verNum, _ = parseVersionTag(raw)
	}

	if phase != "" && phase != "stable" && !strings.Contains(verNum, phase) {
		verNum = fmt.Sprintf("%s-%s", verNum, strings.ToLower(phase))
	}

	if isMono && !strings.HasSuffix(verNum, "_mono") {
		verNum += "_mono"
	}

	if phase == "" {
		phase = releaseType
	}

	return verNum, releaseType, phase
}

// CmdList shows locally installed Godot versions.
func CmdList() {
	godotDir := GetGodotDir()
	if _, err := os.Stat(godotDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Godot installation directory '%s' does not exist.\n", godotDir)
		os.Exit(1)
	}

	installed, err := GetInstalledVersions(godotDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(installed) == 0 {
		fmt.Printf("No Godot versions installed in %s.\n", godotDir)
		return
	}

	activeVersion := GetActiveVersion()

	fmt.Printf("\nInstalled versions in %s:\n\n", godotDir)
	table := newTable([]string{"VERSION", "TYPE", "FULL NAME"})

	for _, iv := range installed {
		verNum, releaseType, phase := parseInstalledVersion(iv.Version)
		isActive := activeVersion != "" && (strings.EqualFold(iv.Version, activeVersion) || MatchesTokens(iv.Version, activeVersion))

		versionDisplay := "  " + verNum
		if isActive {
			versionDisplay = "* " + verNum
		}

		color := colorize(releaseType)
		table.Append([]string{
			color + versionDisplay + resetColor,
			color + strings.ToUpper(phase) + resetColor,
			iv.Version,
		})
	}

	table.Render()
}
