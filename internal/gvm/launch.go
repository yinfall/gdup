package gvm

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/user/gvm/internal/sysutil"
)

// CmdLaunch resolves the active Godot version and forwards all arguments to it.
func CmdLaunch() {
	godotDir := GetGodotDir()
	if _, err := os.Stat(godotDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Godot installation directory '%s' does not exist.\n", godotDir)
		os.Exit(1)
	}

	// Read .gvmrc config
	configPath := FindGvmConfig()
	var version string
	if configPath != "" {
		cfg, err := ReadGvmConfig(configPath)
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

	exeExt := ExeExtension()
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
		normVersion := NormalizeVersionTag(version)

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

		var guiMatches, consoleMatches []string
		for _, m := range matches {
			if strings.Contains(strings.ToLower(m), "_console") {
				consoleMatches = append(consoleMatches, m)
			} else {
				guiMatches = append(guiMatches, m)
			}
		}

		if len(guiMatches) > 0 {
			selectedExe = LatestByVersion(guiMatches)
		} else {
			selectedExe = LatestByVersion(consoleMatches)
		}
	} else {
		var guiBinaries, consoleBinaries []string
		for _, b := range binaries {
			if strings.Contains(strings.ToLower(b), "_console") {
				consoleBinaries = append(consoleBinaries, b)
			} else {
				guiBinaries = append(guiBinaries, b)
			}
		}

		if len(guiBinaries) > 0 {
			selectedExe = LatestByVersion(guiBinaries)
		} else {
			selectedExe = LatestByVersion(consoleBinaries)
		}
	}

	if selectedExe == "" {
		fmt.Fprintln(os.Stderr, "Error: Could not find or access a suitable Godot executable.")
		os.Exit(1)
	}

	exePath := filepath.Join(godotDir, selectedExe)

	cmd := exec.Command(exePath, os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if !strings.Contains(strings.ToLower(selectedExe), "_console") {
		sysutil.SetGUIProcessAttrs(cmd)
	}

	// Ignore signals so the parent process doesn't exit immediately on Ctrl+C.
	// This allows the child process to handle the signal and exit gracefully,
	// preventing terminal state corruption.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
	}()

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "Error: Failed to execute Godot binary: %v\n", err)
		os.Exit(1)
	}
}
