package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/user/gvm/internal/gvmcore"

	"github.com/user/gvm/internal/sysutil"
)

func main() {
	// Attach to parent console (needed when compiled with -H windowsgui)
	sysutil.AttachParentConsole()

	godotDir := gvmcore.GetGodotDir()
	if _, err := os.Stat(godotDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Godot installation directory '%s' does not exist.\n", godotDir)
		os.Exit(1)
	}

	// Read .gvm config
	configPath := gvmcore.FindGvmConfig()
	var version string
	if configPath != "" {
		cfg, err := gvmcore.ReadGvmConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to parse configuration file '%s': %v\n", configPath, err)
		} else {
			version = cfg.Version
		}
	}

	// List all executables in godot-versions directory
	entries, err := os.ReadDir(godotDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to list directory '%s': %v\n", godotDir, err)
		os.Exit(1)
	}

	exeExt := gvmcore.ExeExtension()
	var binaries []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(strings.ToLower(name), exeExt) && !e.IsDir() {
			binaries = append(binaries, name)
		}
	}

	if len(binaries) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No executables found in '%s'.\n", godotDir)
		os.Exit(1)
	}

	var selectedExe string

	if version != "" {
		normVersion := strings.TrimPrefix(version, "v")
		normVersion = strings.ToLower(normVersion)

		// Filter binaries matching the version
		var matches []string
		for _, b := range binaries {
			if strings.Contains(strings.ToLower(b), normVersion) {
				matches = append(matches, b)
			}
		}

		if len(matches) == 0 {
			fmt.Fprintf(os.Stderr, "Error: No Godot binary matching version '%s' found in '%s'.\n", version, godotDir)
			fmt.Fprintln(os.Stderr, "Available binaries in directory:")
			for _, b := range binaries {
				fmt.Fprintf(os.Stderr, "  - %s\n", b)
			}
			os.Exit(1)
		}

		// Prefer GUI (non-console) binary for interactive use
		var guiMatches, consoleMatches []string
		for _, m := range matches {
			if strings.Contains(strings.ToLower(m), "_console") {
				consoleMatches = append(consoleMatches, m)
			} else {
				guiMatches = append(guiMatches, m)
			}
		}

		if len(guiMatches) > 0 {
			selectedExe = latestByVersion(guiMatches)
		} else {
			selectedExe = latestByVersion(consoleMatches)
		}
	} else {
		// No version specified, fallback to latest
		var guiBinaries, consoleBinaries []string
		for _, b := range binaries {
			if strings.Contains(strings.ToLower(b), "_console") {
				consoleBinaries = append(consoleBinaries, b)
			} else {
				guiBinaries = append(guiBinaries, b)
			}
		}

		if len(guiBinaries) > 0 {
			selectedExe = latestByVersion(guiBinaries)
		} else {
			selectedExe = latestByVersion(consoleBinaries)
		}
	}

	if selectedExe == "" {
		fmt.Fprintln(os.Stderr, "Error: Could not find or access a suitable Godot executable.")
		os.Exit(1)
	}

	exePath := filepath.Join(godotDir, selectedExe)

	// Launch Godot: resolve version, then directly forward to the real binary.
	// stdio is passed through — the shim does not sit in the data path.
	cmd := exec.Command(exePath, os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if !strings.Contains(strings.ToLower(selectedExe), "_console") {
		sysutil.SetGUIProcessAttrs(cmd)
	}

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "Error: Failed to execute Godot binary: %v\n", err)
		os.Exit(1)
	}
}

func latestByVersion(files []string) string {
	sort.Slice(files, func(i, j int) bool {
		vi := gvmcore.ParseVersionFromFilename(files[i])
		vj := gvmcore.ParseVersionFromFilename(files[j])
		return compareVersionInfo(vi, vj) < 0
	})
	if len(files) > 0 {
		return files[len(files)-1]
	}
	return ""
}

func compareVersionInfo(a, b gvmcore.VersionInfo) int {
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
