package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/yinfall/gdup/internal/gdup"
	"github.com/yinfall/gdup/internal/sysutil"
)

func main() {
	// Attach to parent console (needed when compiled with -H windowsgui)
	sysutil.AttachParentConsole()

	exeName := filepath.Base(os.Args[0])
	exeName = strings.TrimSuffix(strings.ToLower(exeName), ".exe")

	// Fat Binary: Shim mode
	if exeName == "godot" {
		gdup.CmdLaunch()
		return
	}

	// Act as the version manager CLI "gdup"
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	first := strings.ToLower(os.Args[1])

	switch first {
	case "--help", "-h", "help":
		printHelp()
	case "--version", "-v":
		fmt.Printf("gdup %s (%s/%s)\n", gdup.GDUPVersion, runtime.GOOS, runtime.GOARCH)
	case "list", "ls":
		gdup.CmdList()
	case "releases", "release":
		showAll := false
		update := false
		for _, arg := range os.Args[2:] {
			if arg == "-a" || arg == "--all" {
				showAll = true
			} else if arg == "-u" || arg == "--update" {
				update = true
			}
		}
		gdup.CmdReleases(showAll, update)
	case "install":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: Please specify the version to install. E.g.: gdup install 4.3")
			os.Exit(1)
		}
		gdup.CmdInstall(os.Args[2])
	case "use":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: Please specify the version to use. E.g.: gdup use 4.3")
			os.Exit(1)
		}
		gdup.CmdUse(os.Args[2])
	case "uninstall", "remove":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: Please specify the version to uninstall. E.g.: gdup uninstall 4.3")
			os.Exit(1)
		}
		force := len(os.Args) > 3 && (os.Args[3] == "-y" || os.Args[3] == "--yes")
		gdup.CmdUninstall(os.Args[2], force)
	case "godot":
		// Shift args so CmdLaunch receives the engine arguments correctly
		// E.g., "gdup godot --editor" becomes "godot --editor"
		os.Args = append([]string{"godot"}, os.Args[2:]...)
		gdup.CmdLaunch()
	case "shim":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: Please specify 'install' or 'remove' (e.g., gdup shim install)")
			os.Exit(1)
		}
		gdup.CmdShim(os.Args[2])
	case "license", "licenses":
		gdup.CmdLicense()
	default:
		fmt.Fprintf(os.Stderr, "gdup: unknown command '%s'\nRun 'gdup help' for usage.\n", first)
		os.Exit(1)
	}
}

func printHelp() {
	godotDir := gdup.GetGodotDir()
	fmt.Printf(`Godot Version Manager (GDUP) %s

Usage:
  gdup install <version>           Download and install a specific Godot version (e.g. 4.3)
  gdup list | ls                   Show locally installed Godot versions in %s
  gdup releases [-a] [-u]          List available releases from GitHub (-a for all, -u to force update cache)
  gdup use <version>               Set the active Godot version for the current directory
  gdup uninstall <version> [-y]    Uninstall a locally installed version
  gdup godot [args...]             Run the active Godot version (similar to fvm flutter)
  
Shim Management:
  gdup shim install                Enable transparent proxying (allows you to just type 'godot')
  gdup shim remove                 Disable transparent proxying (removes the shim executable)

Other:
  gdup license                     View third-party open source licenses

  gdup --version                   Show GDUP version
`, gdup.GDUPVersion, godotDir)
}
