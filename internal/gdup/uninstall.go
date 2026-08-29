package gdup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CmdUninstall removes a locally installed Godot version.
func CmdUninstall(version string, force bool) {
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
		fmt.Fprintf(os.Stderr, "Error: Version '%s' is not installed.\n", version)
		os.Exit(1)
	}

	// Warn if active
	activeVersion := GetActiveVersion()
	if VersionMatches(matched.Version, activeVersion) {
		configPath := FindGvmConfig()
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
