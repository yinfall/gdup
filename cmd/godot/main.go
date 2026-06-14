package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/user/gvm/internal/gvm"
	"github.com/user/gvm/internal/sysutil"
)

// GVM sub-commands that are handled by gvm logic.
// Any other argument is forwarded to the real Godot binary.
var gvmCommands = map[string]bool{
	"install":   true,
	"uninstall": true,
	"remove":    true,
	"use":       true,
	"list":      true,
	"ls":        true,
	"releases":  true,
	"release":   true,
}

func main() {
	// Attach to parent console (needed when compiled with -H windowsgui)
	sysutil.AttachParentConsole()

	if len(os.Args) < 2 {
		// No arguments: launch Godot editor
		gvm.CmdLaunch()
		return
	}

	first := strings.ToLower(os.Args[1])

	switch {
	case first == "--help" || first == "-h" || first == "help":
		printHelp()
	case first == "--gvm-version":
		fmt.Printf("gvm %s (%s/%s)\n", gvm.GVMVersion, runtime.GOOS, runtime.GOARCH)
	case gvmCommands[first]:
		handleGVMCommand(first)
	default:
		// Not a gvm command: forward everything to Godot
		gvm.CmdLaunch()
	}
}

func handleGVMCommand(cmd string) {
	switch cmd {
	case "list", "ls":
		gvm.CmdList()
	case "releases", "release":
		showAll := false
		if len(os.Args) > 2 && (os.Args[2] == "-a" || os.Args[2] == "--all") {
			showAll = true
		}
		gvm.CmdReleases(showAll)
	case "install":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: Please specify the version to install. E.g.: godot install 4.6.3")
			os.Exit(1)
		}
		gvm.CmdInstall(os.Args[2])
	case "use":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: Please specify the version to use. E.g.: godot use 4.6.3")
			os.Exit(1)
		}
		gvm.CmdUse(os.Args[2])
	case "uninstall", "remove":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: Please specify the version to uninstall. E.g.: godot uninstall 4.6.3")
			os.Exit(1)
		}
		force := len(os.Args) > 3 && (os.Args[3] == "-y" || os.Args[3] == "--yes")
		gvm.CmdUninstall(os.Args[2], force)
	}
}

func printHelp() {
	godotDir := gvm.GetGodotDir()
	fmt.Printf(`Godot Version Manager (GVM) %s

Usage:
  godot                             Launch Godot editor (forwards to the active version)
  godot [flags...]                  Forward flags to Godot (e.g. godot --version, godot --editor)
  godot install <version>           Download and install a specific Godot version (e.g. 4.6.3)
  godot list | ls                   Show locally installed Godot versions in %s
  godot releases [-a]               List available releases from GitHub (use -a/--all for pre-releases)
  godot use <version>               Set the active Godot version for the current directory
  godot uninstall <version> [-y]    Uninstall a locally installed version (-y to skip confirmation)
  godot --gvm-version               Show GVM version

Example:
  godot install 4.6.3
  godot use 4.6.3
  godot --editor
  godot --version
`, gvm.GVMVersion, godotDir)
}
