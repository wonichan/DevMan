package versionmanager

import (
	"fmt"
	"testing"
)

func TestUninstallAllowsDevManVersion(t *testing.T) {
	env := newFakeEnvironment()
	version := uninstallTestVersion(SourceDevMan, DeletePolicyDirect, true)
	reg := newFakeVersionRegistry([]ManagedVersion{version})

	result, err := NewService(reg, env).UninstallVersion(version, false)
	if err != nil {
		t.Fatalf("UninstallVersion failed: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
	if result.Message != "version uninstalled" {
		t.Fatalf("Message = %q", result.Message)
	}
	if result.ToolKey != version.ToolKey {
		t.Fatalf("ToolKey = %q", result.ToolKey)
	}
	if result.Version != version.Version {
		t.Fatalf("Version = %q", result.Version)
	}
	if len(result.AffectedPaths) != 1 || result.AffectedPaths[0] != version.InstallPath {
		t.Fatalf("AffectedPaths = %#v", result.AffectedPaths)
	}
	if len(env.removedPaths) != 1 || env.removedPaths[0] != version.InstallPath {
		t.Fatalf("removed paths = %#v", env.removedPaths)
	}
	if len(reg.deletedIDs) != 1 || reg.deletedIDs[0] != version.ID {
		t.Fatalf("deleted IDs = %#v", reg.deletedIDs)
	}
}

func TestUninstallReturnsErrorWhenRegistryDeleteFailsAfterRemoval(t *testing.T) {
	env := newFakeEnvironment()
	version := uninstallTestVersion(SourceDevMan, DeletePolicyDirect, true)
	reg := newFakeVersionRegistry([]ManagedVersion{version})
	reg.deleteErr = fmt.Errorf("delete row failed")

	_, err := NewService(reg, env).UninstallVersion(version, false)
	if err == nil {
		t.Fatal("expected registry delete error")
	}
	if err.Error() != "delete row failed" {
		t.Fatalf("error = %q", err)
	}
	if len(env.removedPaths) != 1 || env.removedPaths[0] != version.InstallPath {
		t.Fatalf("removed paths = %#v", env.removedPaths)
	}
	if len(reg.deletedIDs) != 1 || reg.deletedIDs[0] != version.ID {
		t.Fatalf("deleted IDs = %#v", reg.deletedIDs)
	}
}

func TestUninstallBlocksDefaultAndActiveVersions(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*ManagedVersion)
	}{
		{
			name: "default",
			mutate: func(version *ManagedVersion) {
				version.IsDefault = true
			},
		},
		{
			name: "active",
			mutate: func(version *ManagedVersion) {
				version.IsActive = true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newFakeEnvironment()
			version := uninstallTestVersion(SourceDevMan, DeletePolicyDirect, true)
			tt.mutate(&version)

			_, err := NewService(nil, env).UninstallVersion(version, false)
			if err == nil {
				t.Fatal("expected active/default guard error")
			}
			if err.Error() != "active/default versions must be switched before deletion" {
				t.Fatalf("error = %q", err)
			}
			if len(env.removedPaths) != 0 {
				t.Fatalf("removed paths = %#v", env.removedPaths)
			}
		})
	}
}

func TestUninstallBlocksExternalWithoutForce(t *testing.T) {
	env := newFakeEnvironment()
	version := uninstallTestVersion(SourceExternal, DeletePolicyForceRequired, false)

	_, err := NewService(nil, env).UninstallVersion(version, false)
	if err == nil {
		t.Fatal("expected external force confirmation error")
	}
	if err.Error() != "external version requires force confirmation" {
		t.Fatalf("error = %q", err)
	}
	if len(env.removedPaths) != 0 {
		t.Fatalf("removed paths = %#v", env.removedPaths)
	}
}

func TestUninstallBlocksVersionManagerOwned(t *testing.T) {
	env := newFakeEnvironment()
	version := uninstallTestVersion(SourceVersionManager, DeletePolicyBlocked, false)

	_, err := NewService(nil, env).UninstallVersion(version, true)
	if err == nil {
		t.Fatal("expected version manager ownership error")
	}
	if err.Error() != "version manager owned versions cannot be deleted by DevMan" {
		t.Fatalf("error = %q", err)
	}
	if len(env.removedPaths) != 0 {
		t.Fatalf("removed paths = %#v", env.removedPaths)
	}
}

func TestUninstallAllowsExternalWithForce(t *testing.T) {
	env := newFakeEnvironment()
	version := uninstallTestVersion(SourceExternal, DeletePolicyForceRequired, false)

	result, err := NewService(nil, env).UninstallVersion(version, true)
	if err != nil {
		t.Fatalf("UninstallVersion failed: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
	if len(env.removedPaths) != 1 || env.removedPaths[0] != version.InstallPath {
		t.Fatalf("removed paths = %#v", env.removedPaths)
	}
}

func TestUninstallRejectsUnknownSource(t *testing.T) {
	env := newFakeEnvironment()
	version := uninstallTestVersion(VersionSource("unknown"), DeletePolicyDirect, true)

	_, err := NewService(nil, env).UninstallVersion(version, true)
	if err == nil {
		t.Fatal("expected unknown source error")
	}
	if err.Error() != "unknown version source: unknown" {
		t.Fatalf("error = %q", err)
	}
	if len(env.removedPaths) != 0 {
		t.Fatalf("removed paths = %#v", env.removedPaths)
	}
}

func TestUninstallRemoveAllErrorPropagates(t *testing.T) {
	env := newFakeEnvironment()
	env.removeErr = fmt.Errorf("remove failed")
	version := uninstallTestVersion(SourceDevMan, DeletePolicyDirect, true)

	_, err := NewService(nil, env).UninstallVersion(version, false)
	if err == nil {
		t.Fatal("expected remove error")
	}
	if err.Error() != "remove failed" {
		t.Fatalf("error = %q", err)
	}
	if len(env.removedPaths) != 1 || env.removedPaths[0] != version.InstallPath {
		t.Fatalf("removed paths = %#v", env.removedPaths)
	}
}

func TestUninstallRejectsNonMutableEnvironment(t *testing.T) {
	version := uninstallTestVersion(SourceDevMan, DeletePolicyDirect, true)

	_, err := NewService(nil, readOnlyEnvironment{}).UninstallVersion(version, false)
	if err == nil {
		t.Fatal("expected deletion support error")
	}
	if err.Error() != "environment does not support deletion" {
		t.Fatalf("error = %q", err)
	}
}

func TestUninstallRejectsUnsafeInstallPathBeforeRemoval(t *testing.T) {
	tests := []struct {
		name        string
		installPath string
		wantError   string
	}{
		{name: "blank", installPath: "", wantError: "invalid install path: required"},
		{name: "relative", installPath: `go1.26`, wantError: "invalid install path: must be absolute"},
		{name: "drive root", installPath: `D:\`, wantError: "invalid install path: must not be drive root"},
		{name: "quote", installPath: `D:\production\go"1.26`, wantError: "invalid install path: contains quote"},
		{name: "high-level parent", installPath: `D:\production`, wantError: `invalid delete path: D:\production does not look like a go install root`},
		{name: "windows root", installPath: `C:\Windows`, wantError: `invalid delete path: C:\Windows is protected`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newFakeEnvironment()
			version := uninstallTestVersion(SourceDevMan, DeletePolicyDirect, true)
			version.InstallPath = tt.installPath

			_, err := NewService(nil, env).UninstallVersion(version, false)
			if err == nil {
				t.Fatal("expected unsafe install path error")
			}
			if err.Error() != tt.wantError {
				t.Fatalf("error = %q, want %q", err, tt.wantError)
			}
			if len(env.removedPaths) != 0 {
				t.Fatalf("removed paths = %#v", env.removedPaths)
			}
		})
	}
}

func TestUninstallAllowsToolVersionLookingDeletePath(t *testing.T) {
	env := newFakeEnvironment()
	version := uninstallTestVersion(SourceDevMan, DeletePolicyDirect, true)
	version.InstallPath = `D:\production\go1.25.0`

	_, err := NewService(nil, env).UninstallVersion(version, false)
	if err != nil {
		t.Fatalf("UninstallVersion failed: %v", err)
	}
	if len(env.removedPaths) != 1 || env.removedPaths[0] != version.InstallPath {
		t.Fatalf("removed paths = %#v", env.removedPaths)
	}
}

func uninstallTestVersion(source VersionSource, deletePolicy DeletePolicy, canDelete bool) ManagedVersion {
	return ManagedVersion{
		ID:           42,
		ToolKey:      "go",
		Version:      "1.26.0",
		InstallPath:  `D:\production\go1.26.0`,
		Source:       source,
		CanDelete:    canDelete,
		DeletePolicy: deletePolicy,
	}
}
