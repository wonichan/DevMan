//go:build windows

package migrator

import (
	"os/exec"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

func updateWindowsPath(oldPath, newPath string) error {
	// Update User PATH
	key, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer key.Close()

	path, _, err := key.GetStringValue("Path")
	if err != nil {
		return err
	}

	updated := strings.ReplaceAll(path, oldPath, newPath)
	if err := key.SetStringValue("Path", updated); err != nil {
		return err
	}

	// Broadcast environment change
	return broadcastEnvChange()
}

func broadcastEnvChange() error {
	hwnd := uintptr(0xffff) // HWND_BROADCAST
	msg := uintptr(0x1A)    // WM_SETTINGCHANGE
	// Use SendNotifyMessageW
	user32 := syscall.NewLazyDLL("user32.dll")
	proc := user32.NewProc("SendNotifyMessageW")
	_, _, err := proc.Call(hwnd, msg, 0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("Environment"))))
	if err != nil && err != syscall.Errno(0) {
		return err
	}
	return nil
}

func createWindowsJunction(oldPath, newPath string) error {
	// Use mklink /J to create directory junction
	cmd := exec.Command("cmd", "/c", "mklink", "/J", oldPath, newPath)
	return cmd.Run()
}
