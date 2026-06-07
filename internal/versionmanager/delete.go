package versionmanager

import "fmt"

func deleteAllowed(version ManagedVersion, forceExternal bool) error {
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

func (s *Service) UninstallVersion(version ManagedVersion, forceDeleteExternal bool) (*VersionOperationResult, error) {
	if err := deleteAllowed(version, forceDeleteExternal); err != nil {
		return nil, err
	}
	mutable, ok := s.env.(MutableEnvironment)
	if !ok {
		return nil, fmt.Errorf("environment does not support deletion")
	}
	if err := validateInstallPath(version.InstallPath); err != nil {
		return nil, err
	}
	if err := mutable.RemoveAll(version.InstallPath); err != nil {
		return nil, err
	}
	return &VersionOperationResult{
		Success:       true,
		Message:       "version uninstalled",
		ToolKey:       version.ToolKey,
		Version:       version.Version,
		AffectedPaths: []string{version.InstallPath},
	}, nil
}
