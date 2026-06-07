//go:build windows

package versionmanager

import (
	"strings"

	"golang.org/x/sys/windows/registry"
)

const userEnvironmentPath = `Environment`

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
		return envKey.SetStringValue("Path", entry)
	}
	return envKey.SetStringValue("Path", entry+";"+current)
}

func splitPathEntries(value string) []string {
	if value == "" {
		return nil
	}
	return strings.Split(value, ";")
}
