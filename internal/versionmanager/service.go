package versionmanager

import (
	"fmt"
	"path/filepath"
	"strings"
)

type VersionRegistry interface {
	ListToolVersions(toolKey string) ([]ManagedVersion, error)
	SaveToolVersion(v *ManagedVersion) error
	GetInstallStrategy(toolKey string) (*InstallStrategy, error)
	SaveInstallStrategy(strategy InstallStrategy) error
}

type Service struct {
	reg VersionRegistry
	env Environment
}

func NewService(reg VersionRegistry, env Environment) *Service {
	if env == nil {
		env = RealEnvironment{}
	}
	return &Service{reg: reg, env: env}
}

func (s *Service) ListToolVersions() ([]ToolVersionState, error) {
	var states []ToolVersionState
	for _, tool := range SupportedTools() {
		state := ToolVersionState{
			ToolKey:         tool.Key,
			Name:            tool.Name,
			LocalVersions:   []ManagedVersion{},
			ManagerConflict: s.DetectVersionManager(tool.Key),
		}
		if s.reg != nil {
			versions, err := s.reg.ListToolVersions(tool.Key)
			if err != nil {
				return nil, err
			}
			state.LocalVersions = versions
			for i := range versions {
				if versions[i].IsDefault {
					state.CurrentDefault = &versions[i]
					break
				}
			}
		}
		states = append(states, state)
	}
	return states, nil
}

func (s *Service) PreviewVersionInstall(toolKey string, version string) (*VersionInstallPlan, error) {
	if _, ok := ToolByKey(toolKey); !ok {
		return nil, fmt.Errorf("unsupported tool: %s", toolKey)
	}
	if conflict := s.DetectVersionManager(toolKey); conflict != nil && conflict.Detected {
		return nil, fmt.Errorf("%s is managed by %s; DevMan will not take over this tool", toolKey, conflict.Manager)
	}
	return ResolveInstallRoot(s.env, toolKey, version)
}

func (s *Service) SwitchVersion(version ManagedVersion) (*VersionOperationResult, error) {
	tool, ok := ToolByKey(version.ToolKey)
	if !ok {
		return nil, fmt.Errorf("unsupported tool: %s", version.ToolKey)
	}
	if conflict := s.DetectVersionManager(version.ToolKey); conflict != nil && conflict.Detected {
		return nil, fmt.Errorf("%s is managed by %s; DevMan will not take over this tool", version.ToolKey, conflict.Manager)
	}
	mutable, ok := s.env.(MutableEnvironment)
	if !ok {
		return nil, fmt.Errorf("environment does not support mutation")
	}

	exeDir := mutable.ExecutableDir()
	if strings.TrimSpace(exeDir) == "" {
		return nil, fmt.Errorf("executable directory is required")
	}

	if err := validateInstallPath(version.InstallPath); err != nil {
		return nil, err
	}
	targets, err := ShimTargets(tool.Key, version.InstallPath)
	if err != nil {
		return nil, err
	}
	if err := validateShimTargets(targets); err != nil {
		return nil, err
	}
	primaryTarget, ok := primaryShimTarget(tool, targets)
	if !ok {
		return nil, fmt.Errorf("primary shim target not found for %s", tool.Key)
	}
	if !mutable.FileExists(primaryTarget) {
		return nil, fmt.Errorf("expected executable not found: %s", primaryTarget)
	}

	verificationOutput, err := mutable.Run(primaryTarget, tool.VersionArgs...)
	if err != nil {
		return nil, err
	}
	verificationCommand := strings.TrimSpace(strings.Join(append([]string{primaryTarget}, tool.VersionArgs...), " "))

	shimDir := filepath.Join(exeDir, "shims")
	if err := mutable.MkdirAll(shimDir, 0755); err != nil {
		return nil, err
	}

	affectedPaths := []string{shimDir}
	for shimName, target := range targets {
		shimPath := filepath.Join(shimDir, shimName)
		shim, err := GenerateShim(target)
		if err != nil {
			return nil, err
		}
		if err := mutable.WriteFile(shimPath, []byte(shim), 0755); err != nil {
			return nil, err
		}
		affectedPaths = append(affectedPaths, shimPath)
	}

	if err := mutable.SetUserEnv("DEVMAN_HOME", exeDir); err != nil {
		return nil, err
	}
	if err := mutable.SetUserEnv(tool.EnvVar, version.InstallPath); err != nil {
		return nil, err
	}
	const shimPathEntry = `%DEVMAN_HOME%\shims`
	if err := mutable.EnsureUserPathEntry(shimPathEntry); err != nil {
		return nil, err
	}
	if err := s.persistSwitchedVersionState(version); err != nil {
		return nil, err
	}

	return &VersionOperationResult{
		Success:       true,
		Message:       fmt.Sprintf("switched %s to %s", tool.Key, version.Version),
		ToolKey:       tool.Key,
		Version:       version.Version,
		AffectedPaths: affectedPaths,
		AffectedEnvironment: map[string]string{
			"DEVMAN_HOME": exeDir,
			tool.EnvVar:   version.InstallPath,
			"Path":        shimPathEntry,
		},
		RollbackAvailable:   true,
		VerificationCommand: verificationCommand,
		VerificationOutput:  verificationOutput,
	}, nil
}

func (s *Service) persistSwitchedVersionState(selected ManagedVersion) error {
	if s.reg == nil {
		return nil
	}
	versions, err := s.reg.ListToolVersions(selected.ToolKey)
	if err != nil {
		return err
	}
	for _, version := range versions {
		version.IsDefault = version.ID == selected.ID
		version.IsActive = version.ID == selected.ID
		if err := s.reg.SaveToolVersion(&version); err != nil {
			return err
		}
	}
	return nil
}

func validateInstallPath(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("invalid install path: required")
	}
	if strings.Contains(path, `"`) {
		return fmt.Errorf("invalid install path: contains quote")
	}
	if !filepath.IsAbs(path) {
		return fmt.Errorf("invalid install path: must be absolute")
	}
	clean := filepath.Clean(path)
	volume := filepath.VolumeName(clean)
	if volume != "" && strings.EqualFold(clean, volume+string(filepath.Separator)) {
		return fmt.Errorf("invalid install path: must not be drive root")
	}
	if filepath.Dir(clean) == clean {
		return fmt.Errorf("invalid install path: must not be filesystem root")
	}
	return nil
}

func validateShimTargets(targets map[string]string) error {
	for _, target := range targets {
		if strings.Contains(target, `"`) {
			return fmt.Errorf("shim target contains quote: %s", target)
		}
	}
	return nil
}

func primaryShimTarget(tool ToolDefinition, targets map[string]string) (string, bool) {
	primaryShim := strings.TrimSuffix(tool.PrimaryExe, filepath.Ext(tool.PrimaryExe)) + ".cmd"
	target, ok := targets[primaryShim]
	return target, ok
}

func (s *Service) DetectVersionManager(toolKey string) *VersionManagerConflict {
	return DetectVersionManagerConflict(s.env, toolKey)
}
