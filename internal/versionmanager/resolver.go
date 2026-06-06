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
	version = strings.TrimSpace(version)
	if version == "" {
		return nil, fmt.Errorf("version is required")
	}
	if !isSafeVersion(version) {
		return nil, fmt.Errorf("invalid version: %s", version)
	}

	if value := strings.TrimSpace(env.Getenv(tool.EnvVar)); value != "" {
		if plan := planFromExistingRoot(env, tool, version, value, fmt.Sprintf("based on %s=%s", tool.EnvVar, value)); plan != nil {
			return plan, nil
		}
	}

	if exe := env.LookPath(strings.TrimSuffix(tool.PrimaryExe, filepath.Ext(tool.PrimaryExe))); exe != "" {
		root := filepath.Dir(exe)
		if strings.EqualFold(filepath.Base(root), "bin") || strings.EqualFold(filepath.Base(root), "cmd") {
			root = filepath.Dir(root)
		}
		if plan := planFromExistingRoot(env, tool, version, root, fmt.Sprintf("based on PATH executable %s", exe)); plan != nil {
			return plan, nil
		}
	}

	return nil, fmt.Errorf("cannot infer install root for %s", toolKey)
}

func planFromExistingRoot(env Environment, tool ToolDefinition, version string, existingRoot string, reason string) *VersionInstallPlan {
	existingRoot, ok := normalizeExistingRoot(existingRoot)
	if !ok {
		return nil
	}
	if !env.DirExists(existingRoot) {
		return nil
	}

	parent := filepath.Dir(existingRoot)
	targetName := targetDirName(tool.Key, version)
	target := filepath.Join(parent, targetName)
	return &VersionInstallPlan{
		ToolKey:        tool.Key,
		Version:        version,
		TargetDir:      target,
		ExtractedDir:   target,
		WillOverwrite:  env.DirExists(target),
		ResolverReason: reason,
		EnvironmentChanges: map[string]string{
			tool.EnvVar: target,
		},
	}
}

func normalizeExistingRoot(existingRoot string) (string, bool) {
	clean := filepath.Clean(strings.TrimSpace(existingRoot))
	if clean == "" || clean == "." || clean == ".." {
		return "", false
	}
	if !filepath.IsAbs(clean) {
		return "", false
	}
	return clean, true
}

func isSafeVersion(version string) bool {
	if strings.Contains(version, "..") {
		return false
	}
	clean := strings.TrimPrefix(version, "v")
	if clean == "" {
		return false
	}
	first := clean[0]
	if first < '0' || first > '9' {
		return false
	}
	for _, r := range clean {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= 'A' && r <= 'Z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		if r == '.' || r == '_' || r == '-' {
			continue
		}
		return false
	}
	return true
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
