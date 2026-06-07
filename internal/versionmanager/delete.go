package versionmanager

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type toolVersionDeleter interface {
	DeleteToolVersion(id int64) error
}

func deleteAllowed(version ManagedVersion, forceExternal bool) error {
	if version.IsDefault || version.IsActive {
		return fmt.Errorf("active/default versions must be switched before deletion")
	}
	switch version.Source {
	case SourceDevMan:
		if version.CanDelete || version.DeletePolicy == DeletePolicyDirect {
			return nil
		}
		return fmt.Errorf("DevMan version is not marked deletable")
	case SourceExternal:
		if forceExternal {
			return nil
		}
		return fmt.Errorf("external version requires force confirmation")
	case SourceVersionManager:
		return fmt.Errorf("version manager owned versions cannot be deleted by DevMan")
	default:
		return fmt.Errorf("unknown version source: %s", version.Source)
	}
}

func validateDeletePath(version ManagedVersion, forceExternal bool) error {
	if err := validateInstallPath(version.InstallPath); err != nil {
		return err
	}

	clean := filepath.Clean(strings.TrimSpace(version.InstallPath))
	if isProtectedDeletePath(clean) {
		return fmt.Errorf("invalid delete path: %s is protected", clean)
	}
	if err := validateToolInstallLeaf(version.ToolKey, clean); err != nil {
		return err
	}
	return nil
}

func isProtectedDeletePath(path string) bool {
	protected := []string{
		os.Getenv("WINDIR"),
		os.Getenv("SystemRoot"),
		os.Getenv("ProgramFiles"),
		os.Getenv("ProgramFiles(x86)"),
		os.Getenv("USERPROFILE"),
		os.TempDir(),
	}
	if cwd, err := os.Getwd(); err == nil {
		protected = append(protected, cwd)
	}

	clean := filepath.Clean(path)
	for _, candidate := range protected {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		if samePath(clean, filepath.Clean(candidate)) {
			return true
		}
	}
	return false
}

func validateToolInstallLeaf(toolKey string, path string) error {
	path = filepath.Clean(strings.TrimSpace(path))
	leaf := filepath.Base(path)
	var patterns []*regexp.Regexp
	switch toolKey {
	case "go":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`(?i)^go$`),
			regexp.MustCompile(`(?i)^go\d+(?:\.\d+){1,3}$`),
		}
	case "node":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`(?i)^node$`),
			regexp.MustCompile(`(?i)^node-v?\d+(?:\.\d+){1,3}$`),
			regexp.MustCompile(`(?i)^nodejs-v?\d+(?:\.\d+){1,3}$`),
		}
	case "bun":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`(?i)^bun$`),
			regexp.MustCompile(`(?i)^bun-v?\d+(?:\.\d+){1,3}$`),
		}
	case "flutter":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`(?i)^flutter$`),
			regexp.MustCompile(`(?i)^flutter-v?\d+(?:\.\d+){1,3}$`),
		}
	default:
		return fmt.Errorf("unsupported tool: %s", toolKey)
	}
	for _, pattern := range patterns {
		if pattern.MatchString(leaf) {
			return nil
		}
	}
	return fmt.Errorf("invalid delete path: %s does not look like a %s install root", path, toolKey)
}

func samePath(a string, b string) bool {
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
}

func (s *Service) UninstallVersion(version ManagedVersion, forceDeleteExternal bool) (*VersionOperationResult, error) {
	if err := deleteAllowed(version, forceDeleteExternal); err != nil {
		return nil, err
	}
	mutable, ok := s.env.(MutableEnvironment)
	if !ok {
		return nil, fmt.Errorf("environment does not support deletion")
	}
	if err := validateDeletePath(version, forceDeleteExternal); err != nil {
		return nil, err
	}
	if err := mutable.RemoveAll(version.InstallPath); err != nil {
		return nil, err
	}
	if deleter, ok := s.reg.(toolVersionDeleter); ok {
		if err := deleter.DeleteToolVersion(version.ID); err != nil {
			return nil, err
		}
	}
	return &VersionOperationResult{
		Success:       true,
		Message:       "version uninstalled",
		ToolKey:       version.ToolKey,
		Version:       version.Version,
		AffectedPaths: []string{version.InstallPath},
	}, nil
}
