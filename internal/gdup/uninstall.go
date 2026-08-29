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
		if MatchesTokens(installed[i].Version, tag) {
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
	if activeVersion != "" && MatchesTokens(matched.Version, activeVersion) {
		configPath := FindGvmConfig()
		fmt.Printf("Warning: '%s' is currently the active version in '%s'.\n", matched.Version, configPath)
	}

	fmt.Println("The following version directory will be removed:")
	fmt.Printf("  - %s\n", filepath.Join(godotDir, matched.Version))

	if !force {
		fmt.Printf("\nUninstall Godot %s? [y/N] ", matched.Version)
		var answer string
		fmt.Scanln(&answer)
		if strings.ToLower(answer) != "y" && strings.ToLower(answer) != "yes" {
			fmt.Println("Aborted.")
			os.Exit(0)
		}
	}

	p := filepath.Join(godotDir, matched.Version)
	if err := os.RemoveAll(p); err != nil {
		fmt.Fprintf(os.Stderr, "Error removing directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully uninstalled Godot %s!\n", matched.Version)
}
