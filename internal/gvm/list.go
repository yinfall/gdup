package gvm

import (
	"fmt"
	"os"
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

	groups := map[string][]string{}
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(strings.ToLower(name), ExeExtension()) && !e.IsDir() {
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

	sort.Slice(result, func(i, j int) bool {
		vi := ParseVersionFromFilename(result[i].Files[0])
		vj := ParseVersionFromFilename(result[j].Files[0])
		return CompareVersionInfo(vi, vj) > 0
	})

	return result, nil
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

	fmt.Printf("Installed versions in %s:\n", godotDir)
	for _, iv := range installed {
		isActive := VersionMatches(iv.Version, activeVersion)

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
