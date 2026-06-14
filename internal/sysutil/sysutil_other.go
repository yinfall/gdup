//go:build !windows

package sysutil

import (
	"os/exec"
)

func AttachParentConsole() {
	// No-op on non-Windows platforms
}

func SetGUIProcessAttrs(cmd *exec.Cmd) {
	// No special handling needed on non-Windows platforms
}
