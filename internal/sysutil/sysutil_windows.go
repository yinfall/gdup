//go:build windows

package sysutil

import (
	"os"
	"os/exec"

	"golang.org/x/sys/windows"
)

var (
	kernel32          = windows.NewLazyDLL("kernel32.dll")
	procAttachConsole = kernel32.NewProc("AttachConsole")
	procGetStdHandle  = kernel32.NewProc("GetStdHandle")
	procGetFileType   = kernel32.NewProc("GetFileType")
)

const (
	FILE_TYPE_PIPE = 0x0003
)

// AttachParentConsole attaches the process to the parent's console.
func AttachParentConsole() {
	if isFileType(os.Stdout.Fd(), FILE_TYPE_PIPE) {
		return
	}

	ret, _, _ := procAttachConsole.Call(^uintptr(0)) // ATTACH_PARENT_PROCESS
	if ret == 0 {
		return
	}

	hStdout, _, _ := procGetStdHandle.Call(^uintptr(0) - 10) // STD_OUTPUT_HANDLE
	hStderr, _, _ := procGetStdHandle.Call(^uintptr(0) - 11) // STD_ERROR_HANDLE

	if hStdout != 0 && hStdout != uintptr(windows.InvalidHandle) {
		os.Stdout = os.NewFile(hStdout, "/dev/stdout")
	}
	if hStderr != 0 && hStderr != uintptr(windows.InvalidHandle) {
		os.Stderr = os.NewFile(hStderr, "/dev/stderr")
	}
}

func SetGUIProcessAttrs(cmd *exec.Cmd) {
	cmd.SysProcAttr = &windows.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NO_WINDOW,
	}
}

func isFileType(fd uintptr, expected uintptr) bool {
	ft, _, _ := procGetFileType.Call(fd)
	return ft == expected
}
