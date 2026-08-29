package gdup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func CmdShim(action string) {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not determine executable path: %v\n", err)
		os.Exit(1)
	}

	dir := filepath.Dir(exePath)
	shimName := "godot" + ExeExtension()
	shimPath := filepath.Join(dir, shimName)

	switch action {
	case "install", "enable":
		if _, err := os.Stat(shimPath); err == nil {
			fmt.Printf("Shim is already installed at: %s\n", shimPath)
			return
		}

		if err := copyFile(exePath, shimPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error installing shim: %v\n", err)
			os.Exit(1)
		}

		if ExeExtension() == "" {
			os.Chmod(shimPath, 0755)
		}

		fmt.Printf("Success! Transparent shim installed at: %s\n", shimPath)
		fmt.Println("You can now directly use the 'godot' command in your terminal.")

	case "uninstall", "remove", "disable":
		if _, err := os.Stat(shimPath); os.IsNotExist(err) {
			fmt.Printf("Shim is not currently installed.\n")
			return
		}

		if err := os.Remove(shimPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error removing shim: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Success! Shim removed from: %s\n", shimPath)

	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown shim action '%s'. Use 'install' or 'remove'.\n", action)
		os.Exit(1)
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
