//go:build windows

package migrator

import (
	"os/exec"
	"strings"
	"syscall"
	"unsafe"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/registry"
)

func updateWindowsPath(oldPath, newPath string) error {
	logrus.WithFields(logrus.Fields{"old_path": oldPath, "new_path": newPath}).Info("updating Windows user PATH")
	// Update User PATH
	key, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.ALL_ACCESS)
	if err != nil {
		logrus.WithError(err).Error("failed to open Windows environment registry key")
		return err
	}
	defer key.Close()

	path, _, err := key.GetStringValue("Path")
	if err != nil {
		logrus.WithError(err).Error("failed to read Windows user PATH")
		return err
	}

	updated := strings.ReplaceAll(path, oldPath, newPath)
	if err := key.SetStringValue("Path", updated); err != nil {
		logrus.WithError(err).Error("failed to write Windows user PATH")
		return err
	}

	// Broadcast environment change
	if err := broadcastEnvChange(); err != nil {
		logrus.WithError(err).Error("failed to broadcast Windows environment change")
		return err
	}
	logrus.Info("Windows user PATH updated")
	return nil
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
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
	if err := cmd.Run(); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"old_path": oldPath, "new_path": newPath}).Error("failed to create Windows junction")
		return err
	}
	logrus.WithFields(logrus.Fields{"old_path": oldPath, "new_path": newPath}).Info("Windows junction created")
	return nil
}
