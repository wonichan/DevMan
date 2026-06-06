package versionmanager

import (
	"fmt"
	"path/filepath"
	"strings"
)

func ResolveInstallRoot(env Environment, toolKey string, version string) (*VersionInstallPlan, error) {
	tool, ok := ToolByKey(toolKey)
	if !ok {
		return nil, fmt.Errorf("unsupported tool: %s", toolKey)
	}
	if strings.TrimSpace(version) == "" {
		return nil, fmt.Errorf("version is required")
	}

	if value := env.Getenv(tool.EnvVar); value != "" {
		return planFromExistingRoot(tool, version, value, fmt.Sprintf("based on %s=%s", tool.EnvVar, value)), nil
	}

	if exe := env.LookPath(strings.TrimSuffix(tool.PrimaryExe, filepath.Ext(tool.PrimaryExe))); exe != "" {
		root := filepath.Dir(exe)
		if strings.EqualFold(filepath.Base(root), "bin") || strings.EqualFold(filepath.Base(root), "cmd") {
			root = filepath.Dir(root)
		}
		return planFromExistingRoot(tool, version, root, fmt.Sprintf("based on PATH executable %s", exe)), nil
	}

	return nil, fmt.Errorf("cannot infer install root for %s", toolKey)
}

func planFromExistingRoot(tool ToolDefinition, version string, existingRoot string, reason string) *VersionInstallPlan {
	parent := filepath.Dir(existingRoot)
	targetName := targetDirName(tool.Key, version)
	target := filepath.Join(parent, targetName)
	return &VersionInstallPlan{
		ToolKey:        tool.Key,
		Version:        version,
		TargetDir:      target,
		ExtractedDir:   target,
		ResolverReason: reason,
		EnvironmentChanges: map[string]string{
			tool.EnvVar: target,
		},
	}
}

func targetDirName(toolKey string, version string) string {
	clean := strings.TrimPrefix(version, "v")
	switch toolKey {
	case "go":
		return "go" + clean
	case "node":
		return "node-v" + clean
	case "bun":
		return "bun-v" + clean
	case "flutter":
		return "flutter-" + clean
	default:
		return toolKey + "-" + clean
	}
}
