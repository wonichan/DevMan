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
	shimDir := filepath.Join(exeDir, "shims")
	if err := mutable.MkdirAll(shimDir, 0755); err != nil {
		return nil, err
	}

	targets, err := ShimTargets(tool.Key, version.InstallPath)
	if err != nil {
		return nil, err
	}
	affectedPaths := []string{shimDir}
	for shimName, target := range targets {
		shimPath := filepath.Join(shimDir, shimName)
		if err := mutable.WriteFile(shimPath, []byte(GenerateShim(target)), 0755); err != nil {
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

	command := strings.TrimSuffix(tool.PrimaryExe, filepath.Ext(tool.PrimaryExe))
	verificationOutput, err := mutable.Run(command, tool.VersionArgs...)
	if err != nil {
		return nil, err
	}
	verificationCommand := strings.TrimSpace(strings.Join(append([]string{command}, tool.VersionArgs...), " "))

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

func (s *Service) DetectVersionManager(toolKey string) *VersionManagerConflict {
	return DetectVersionManagerConflict(s.env, toolKey)
}
