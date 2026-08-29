package gdup

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/user/gdup/internal/sysutil"
)



// CmdLaunch resolves the active Godot version and forwards all arguments to it.
func CmdLaunch() {
	godotDir := GetGodotDir()
	if _, err := os.Stat(godotDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Godot installation directory '%s' does not exist.\n", godotDir)
		os.Exit(1)
	}

	// Read .gduprc config
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

	installed, err := GetInstalledVersions(godotDir)
	if err != nil || len(installed) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No Godot executables found in '%s'.\n", godotDir)
		os.Exit(1)
	}

	var targetVersion *InstalledVersion
	if version != "" {
		var candidates []string
		for _, iv := range installed {
			candidates = append(candidates, iv.Version)
		}

		bestMatch := QueryBestMatch(candidates, version, "")
		if bestMatch != "" {
			for _, iv := range installed {
				if iv.Version == bestMatch {
					targetVersion = &iv
					break
				}
			}
		}
	} else {
		targetVersion = &installed[0]
	}

	if targetVersion == nil {
		fmt.Fprintf(os.Stderr, "Error: No Godot version matching '%s' found in '%s'.\n", version, godotDir)
		os.Exit(1)
	}

	var guiFiles, consoleFiles []string
	for _, f := range targetVersion.Files {
		if strings.Contains(strings.ToLower(f), "_console") {
			consoleFiles = append(consoleFiles, f)
		} else {
			guiFiles = append(guiFiles, f)
		}
	}

	var selectedExe string
	if len(guiFiles) > 0 {
		selectedExe = guiFiles[0]
	} else if len(consoleFiles) > 0 {
		selectedExe = consoleFiles[0]
	}

	if selectedExe == "" {
		fmt.Fprintln(os.Stderr, "Error: Could not find or access a suitable Godot executable.")
		os.Exit(1)
	}

	exePath := filepath.Join(godotDir, targetVersion.Version, selectedExe)

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
