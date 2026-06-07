package versionmanager

import (
	"fmt"
	"path/filepath"
)

func GenerateShim(target string) string {
	return fmt.Sprintf("@echo off\r\n\"%s\" %%*\r\n", target)
}

func ShimTargets(toolKey, installPath string) (map[string]string, error) {
	if _, ok := ToolByKey(toolKey); !ok {
		return nil, fmt.Errorf("unsupported tool: %s", toolKey)
	}

	switch toolKey {
	case "go":
		return map[string]string{
			"go.cmd": filepath.Join(installPath, "bin", "go.exe"),
		}, nil
	case "node":
		return map[string]string{
			"node.cmd": filepath.Join(installPath, "node.exe"),
			"npm.cmd":  filepath.Join(installPath, "npm.cmd"),
			"npx.cmd":  filepath.Join(installPath, "npx.cmd"),
		}, nil
	case "bun":
		return map[string]string{
			"bun.cmd": filepath.Join(installPath, "bun.exe"),
		}, nil
	case "flutter":
		return map[string]string{
			"flutter.cmd": filepath.Join(installPath, "bin", "flutter.bat"),
			"dart.cmd":    filepath.Join(installPath, "bin", "dart.exe"),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported tool: %s", toolKey)
	}
}
