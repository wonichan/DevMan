package versionmanager

import "fmt"

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
	if conflict := s.DetectVersionManager(toolKey); conflict != nil && conflict.Detected {
		return nil, fmt.Errorf("%s is managed by %s; DevMan will not take over this tool", toolKey, conflict.Manager)
	}
	return ResolveInstallRoot(s.env, toolKey, version)
}

func (s *Service) DetectVersionManager(toolKey string) *VersionManagerConflict {
	return DetectVersionManagerConflict(s.env, toolKey)
}
