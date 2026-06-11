package main

import (
	"devman/internal/versionmanager"
	"fmt"

	"github.com/sirupsen/logrus"
)

// ListToolVersions returns supported version-managed tools and known local versions.
func (a *App) ListToolVersions() ([]versionmanager.ToolVersionState, error) {
	if a.versionManager == nil {
		logrus.Error("list tool versions failed: version manager not initialized")
		return nil, fmt.Errorf("version manager not initialized")
	}
	states, err := a.versionManager.ListToolVersions()
	if err != nil {
		logrus.WithError(err).Error("list tool versions failed")
		return nil, err
	}
	logrus.WithField("tool_count", len(states)).Info("list tool versions completed")
	return states, nil
}

// PreviewVersionInstall resolves where a version install would be placed.
func (a *App) PreviewVersionInstall(toolKey string, version string) (*versionmanager.VersionInstallPlan, error) {
	logrus.WithFields(logrus.Fields{"tool_key": toolKey, "version": version}).Info("preview version install requested")
	if a.versionManager == nil {
		logrus.Error("preview version install failed: version manager not initialized")
		return nil, fmt.Errorf("version manager not initialized")
	}
	plan, err := a.versionManager.PreviewVersionInstall(toolKey, version)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"tool_key": toolKey, "version": version}).Warn("preview version install failed")
		return nil, err
	}
	return plan, nil
}

// InstallVersion downloads, extracts, and records a DevMan-managed tool version.
func (a *App) InstallVersion(toolKey string, version string, targetDir string) (*versionmanager.VersionOperationResult, error) {
	logrus.WithFields(logrus.Fields{"tool_key": toolKey, "version": version, "target_dir": targetDir}).Info("install version requested")
	if a.versionManager == nil {
		logrus.Error("install version failed: version manager not initialized")
		return nil, fmt.Errorf("version manager not initialized")
	}
	result, err := a.versionManager.InstallVersion(toolKey, version, targetDir)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"tool_key": toolKey, "version": version, "target_dir": targetDir}).Warn("install version failed")
		return nil, err
	}
	logrus.WithFields(logrus.Fields{"tool_key": toolKey, "version": version, "target_dir": targetDir}).Info("install version completed")
	return result, nil
}

// FetchOfficialVersions retrieves available versions from the tool's official source.
func (a *App) FetchOfficialVersions(toolKey string) (*versionmanager.ToolVersionCatalog, error) {
	logrus.WithField("tool_key", toolKey).Info("fetch official versions requested")
	if a.versionManager == nil {
		logrus.Error("fetch official versions failed: version manager not initialized")
		return nil, fmt.Errorf("version manager not initialized")
	}
	catalog, err := a.versionManager.FetchOfficialVersions(toolKey)
	if err != nil {
		logrus.WithError(err).WithField("tool_key", toolKey).Warn("fetch official versions failed")
		return nil, err
	}
	logrus.WithFields(logrus.Fields{"tool_key": toolKey, "version_count": len(catalog.Versions)}).Info("fetch official versions completed")
	return catalog, nil
}

// SwitchVersion makes a tracked tool version the active user version through DevMan shims.
func (a *App) SwitchVersion(toolKey string, instanceId int64) (*versionmanager.VersionOperationResult, error) {
	logrus.WithFields(logrus.Fields{"tool_key": toolKey, "instance_id": instanceId}).Info("switch version requested")
	if a.versionManager == nil {
		logrus.Error("switch version failed: version manager not initialized")
		return nil, fmt.Errorf("version manager not initialized")
	}
	if a.reg == nil {
		logrus.Error("switch version failed: registry not initialized")
		return nil, fmt.Errorf("registry not initialized")
	}

	versions, err := a.reg.ListToolVersions(toolKey)
	if err != nil {
		logrus.WithError(err).WithField("tool_key", toolKey).Error("switch version failed to list versions")
		return nil, err
	}
	for _, version := range versions {
		if version.ID == instanceId {
			result, err := a.versionManager.SwitchVersion(version)
			if err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{"tool_key": toolKey, "instance_id": instanceId}).Warn("switch version failed")
				return nil, err
			}
			logrus.WithFields(logrus.Fields{"tool_key": toolKey, "instance_id": instanceId}).Info("switch version completed")
			return result, nil
		}
	}
	logrus.WithFields(logrus.Fields{"tool_key": toolKey, "instance_id": instanceId}).Warn("switch version failed: version not found")
	return nil, fmt.Errorf("version not found: %d", instanceId)
}

// UninstallVersion removes a tracked tool version according to its delete policy.
func (a *App) UninstallVersion(instanceId int64, forceDeleteExternal bool) (*versionmanager.VersionOperationResult, error) {
	logrus.WithFields(logrus.Fields{"instance_id": instanceId, "force_delete_external": forceDeleteExternal}).Info("uninstall version requested")
	if a.versionManager == nil {
		logrus.Error("uninstall version failed: version manager not initialized")
		return nil, fmt.Errorf("version manager not initialized")
	}
	if a.reg == nil {
		logrus.Error("uninstall version failed: registry not initialized")
		return nil, fmt.Errorf("registry not initialized")
	}

	for _, tool := range versionmanager.SupportedTools() {
		versions, err := a.reg.ListToolVersions(tool.Key)
		if err != nil {
			logrus.WithError(err).WithField("tool_key", tool.Key).Error("uninstall version failed to list versions")
			return nil, err
		}
		for _, version := range versions {
			if version.ID == instanceId {
				result, err := a.versionManager.UninstallVersion(version, forceDeleteExternal)
				if err != nil {
					logrus.WithError(err).WithFields(logrus.Fields{"tool_key": version.ToolKey, "instance_id": instanceId}).Warn("uninstall version failed")
					return nil, err
				}
				logrus.WithFields(logrus.Fields{"tool_key": version.ToolKey, "instance_id": instanceId}).Info("uninstall version completed")
				return result, nil
			}
		}
	}
	logrus.WithField("instance_id", instanceId).Warn("uninstall version failed: version not found")
	return nil, fmt.Errorf("version not found: %d", instanceId)
}

// DetectVersionManager reports external version-manager ownership for a tool.
func (a *App) DetectVersionManager(toolKey string) *versionmanager.VersionManagerConflict {
	if a.versionManager == nil {
		logrus.WithField("tool_key", toolKey).Warn("detect version manager skipped: version manager not initialized")
		return nil
	}
	return a.versionManager.DetectVersionManager(toolKey)
}
