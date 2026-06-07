//go:build windows

package versionmanager

import (
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const userEnvironmentPath = `Environment`

const (
	hwndBroadcast   = 0xffff
	wmSettingChange = 0x001A
	smtoAbortIfHung = 0x0002
)

var (
	user32                 = windows.NewLazySystemDLL("user32.dll")
	procSendMessageTimeout = user32.NewProc("SendMessageTimeoutW")
)

func setUserEnv(key string, value string) error {
	envKey, err := registry.OpenKey(registry.CURRENT_USER, userEnvironmentPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer envKey.Close()

	return envKey.SetStringValue(key, value)
}

func ensureUserPathEntry(entry string) error {
	envKey, err := registry.OpenKey(registry.CURRENT_USER, userEnvironmentPath, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer envKey.Close()

	current, _, err := envKey.GetStringValue("Path")
	if err != nil && err != registry.ErrNotExist {
		return err
	}

	parts := splitPathEntries(current)
	for _, part := range parts {
		if strings.EqualFold(strings.TrimSpace(part), entry) {
			return nil
		}
	}

	if strings.TrimSpace(current) == "" {
		if err := envKey.SetExpandStringValue("Path", entry); err != nil {
			return err
		}
		broadcastEnvironmentChange()
		return nil
	}
	if err := envKey.SetExpandStringValue("Path", entry+";"+current); err != nil {
		return err
	}
	broadcastEnvironmentChange()
	return nil
}

func splitPathEntries(value string) []string {
	if value == "" {
		return nil
	}
	return strings.Split(value, ";")
}

func broadcastEnvironmentChange() {
	environment, err := syscall.UTF16PtrFromString("Environment")
	if err != nil {
		return
	}
	// Best effort: notify Explorer and future shells without making registry writes depend on UI messaging.
	_, _, _ = procSendMessageTimeout.Call(
		uintptr(hwndBroadcast),
		uintptr(wmSettingChange),
		0,
		uintptr(unsafe.Pointer(environment)),
		uintptr(smtoAbortIfHung),
		uintptr(5000),
		0,
	)
}
