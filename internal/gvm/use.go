package gvm

import (
	"fmt"
	"os"
	"path/filepath"
)

// CmdUse sets the active Godot version for the current directory.
func CmdUse(version string) {
	godotDir := GetGodotDir()
	if _, err := os.Stat(godotDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Godot installation directory '%s' does not exist.\n", godotDir)
		os.Exit(1)
	}

	tag := NormalizeVersion(version)

	installed, err := GetInstalledVersions(godotDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var matched *InstalledVersion
	for i := range installed {
		if VersionMatches(installed[i].Version, tag) {
			matched = &installed[i]
			break
		}
	}

	if matched == nil {
		fmt.Fprintf(os.Stderr, "Error: Version '%s' is not installed locally in %s.\n", version, godotDir)
		fmt.Fprintln(os.Stderr, "Please install it first using: godot install <version>")
		os.Exit(1)
	}

	cwd, _ := os.Getwd()
	configPath := filepath.Join(cwd, ".gvm")
	if err := WriteGvmConfig(configPath, matched.Version); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to write configuration file '%s': %v\n", configPath, err)
		os.Exit(1)
	}

	fmt.Printf("Success: Active Godot version set to '%s' in this directory (%s).\n", matched.Version, configPath)
}
