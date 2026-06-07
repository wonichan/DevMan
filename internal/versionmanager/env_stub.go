//go:build !windows

package versionmanager

import "fmt"

func setUserEnv(key string, value string) error {
	return fmt.Errorf("user environment writes are not supported on this platform")
}

func ensureUserPathEntry(entry string) error {
	return fmt.Errorf("user PATH writes are not supported on this platform")
}
