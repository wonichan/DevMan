package versionmanager

import (
	"path/filepath"
	"testing"
)

func TestSwitchVersionWritesShimsAndEnvironment(t *testing.T) {
	env := newFakeEnvironment()
	env.exeDir = `D:\apps\DevMan`
	env.runOutput = "go version go1.26.0 windows/amd64"

	version := ManagedVersion{
		ID:          42,
		ToolKey:     "go",
		Version:     "1.26.0",
		InstallPath: `D:\production\go1.26`,
	}
	env.files[`D:\production\go1.26\bin\go.exe`] = true

	result, err := NewService(nil, env).SwitchVersion(version)
	if err != nil {
		t.Fatalf("SwitchVersion failed: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
	if env.vars["GOROOT"] != version.InstallPath {
		t.Fatalf("GOROOT = %q", env.vars["GOROOT"])
	}
	if env.vars["DEVMAN_HOME"] != env.exeDir {
		t.Fatalf("DEVMAN_HOME = %q", env.vars["DEVMAN_HOME"])
	}
	if len(env.userPathEntries) != 1 || env.userPathEntries[0] != `%DEVMAN_HOME%\shims` {
		t.Fatalf("user path entries = %#v", env.userPathEntries)
	}

	shimPath := `D:\apps\DevMan\shims\go.cmd`
	expectedShim, err := GenerateShim(`D:\production\go1.26\bin\go.exe`)
	if err != nil {
		t.Fatalf("GenerateShim failed: %v", err)
	}
	if string(env.writes[shimPath]) != expectedShim {
		t.Fatalf("shim content = %q", string(env.writes[shimPath]))
	}
	if env.writePerms[shimPath] != 0755 {
		t.Fatalf("shim perm = %v", env.writePerms[shimPath])
	}
	if env.mkdirs[`D:\apps\DevMan\shims`] != 0755 {
		t.Fatalf("shim dir perm = %v", env.mkdirs[`D:\apps\DevMan\shims`])
	}
	if len(env.runCommands) != 1 {
		t.Fatalf("run commands = %#v", env.runCommands)
	}
	if env.runCommands[0].command != `D:\production\go1.26\bin\go.exe` || len(env.runCommands[0].args) != 1 || env.runCommands[0].args[0] != "version" {
		t.Fatalf("run command = %#v", env.runCommands[0])
	}
	if result.VerificationCommand != `D:\production\go1.26\bin\go.exe version` {
		t.Fatalf("VerificationCommand = %q", result.VerificationCommand)
	}
	if result.VerificationOutput != env.runOutput {
		t.Fatalf("VerificationOutput = %q", result.VerificationOutput)
	}
	if result.AffectedEnvironment["GOROOT"] != version.InstallPath {
		t.Fatalf("AffectedEnvironment GOROOT = %q", result.AffectedEnvironment["GOROOT"])
	}
	if result.AffectedEnvironment["DEVMAN_HOME"] != env.exeDir {
		t.Fatalf("AffectedEnvironment DEVMAN_HOME = %q", result.AffectedEnvironment["DEVMAN_HOME"])
	}
}

func TestSwitchVersionUsesSharedShimPathForEverySupportedTool(t *testing.T) {
	for _, tool := range SupportedTools() {
		t.Run(tool.Key, func(t *testing.T) {
			env := newFakeEnvironment()
			env.exeDir = `D:\apps\DevMan`
			installPath := `D:\production\` + tool.Key + `-2.0.0`
			targets, err := ShimTargets(tool.Key, installPath)
			if err != nil {
				t.Fatalf("ShimTargets failed: %v", err)
			}
			primaryTarget, ok := primaryShimTarget(tool, targets)
			if !ok {
				t.Fatalf("primary shim target not found for %s", tool.Key)
			}
			env.files[primaryTarget] = true
			env.runOutput = "selected"

			result, err := NewService(nil, env).SwitchVersion(ManagedVersion{
				ID:          42,
				ToolKey:     tool.Key,
				Version:     "2.0.0",
				InstallPath: installPath,
			})
			if err != nil {
				t.Fatalf("SwitchVersion failed: %v", err)
			}

			if env.vars[tool.EnvVar] != installPath {
				t.Fatalf("%s = %q, want %q", tool.EnvVar, env.vars[tool.EnvVar], installPath)
			}
			if env.vars["DEVMAN_HOME"] != env.exeDir {
				t.Fatalf("DEVMAN_HOME = %q", env.vars["DEVMAN_HOME"])
			}
			if len(env.userPathEntries) != 1 || env.userPathEntries[0] != `%DEVMAN_HOME%\shims` {
				t.Fatalf("user path entries = %#v", env.userPathEntries)
			}

			for shimName, target := range targets {
				shimPath := filepath.Join(env.exeDir, "shims", shimName)
				expectedShim, err := GenerateShim(target)
				if err != nil {
					t.Fatalf("GenerateShim failed: %v", err)
				}
				if string(env.writes[shimPath]) != expectedShim {
					t.Fatalf("shim %s content = %q", shimName, string(env.writes[shimPath]))
				}
			}
			if result.AffectedEnvironment[tool.EnvVar] != installPath {
				t.Fatalf("AffectedEnvironment %s = %q", tool.EnvVar, result.AffectedEnvironment[tool.EnvVar])
			}
			if result.AffectedEnvironment["Path"] != `%DEVMAN_HOME%\shims` {
				t.Fatalf("AffectedEnvironment Path = %q", result.AffectedEnvironment["Path"])
			}
		})
	}
}

func TestSwitchVersionPersistsSelectedDefaultAndActiveAfterSuccess(t *testing.T) {
	env := newFakeEnvironment()
	env.exeDir = `D:\apps\DevMan`
	env.files[`D:\production\go1.26\bin\go.exe`] = true
	reg := newFakeVersionRegistry([]ManagedVersion{
		{
			ID:          1,
			ToolKey:     "go",
			Version:     "1.25.0",
			InstallPath: `D:\production\go1.25`,
			IsDefault:   true,
			IsActive:    true,
		},
		{
			ID:          2,
			ToolKey:     "go",
			Version:     "1.26.0",
			InstallPath: `D:\production\go1.26`,
		},
	})

	_, err := NewService(reg, env).SwitchVersion(reg.versions[1])
	if err != nil {
		t.Fatalf("SwitchVersion failed: %v", err)
	}

	if len(reg.saved) != 2 {
		t.Fatalf("saved versions = %#v", reg.saved)
	}
	first := reg.savedByID(1)
	if first == nil {
		t.Fatal("version 1 was not saved")
	}
	if first.IsDefault || first.IsActive {
		t.Fatalf("version 1 state = default:%t active:%t", first.IsDefault, first.IsActive)
	}
	selected := reg.savedByID(2)
	if selected == nil {
		t.Fatal("version 2 was not saved")
	}
	if !selected.IsDefault || !selected.IsActive {
		t.Fatalf("version 2 state = default:%t active:%t", selected.IsDefault, selected.IsActive)
	}
}

func TestSwitchVersionVerificationFailureDoesNotMutate(t *testing.T) {
	env := newFakeEnvironment()
	env.exeDir = `D:\apps\DevMan`
	env.files[`D:\production\go1.26\bin\go.exe`] = true
	env.runErr = errFakeRunFailed()

	_, err := NewService(nil, env).SwitchVersion(ManagedVersion{
		ToolKey:     "go",
		Version:     "1.26.0",
		InstallPath: `D:\production\go1.26`,
	})
	if err == nil {
		t.Fatal("expected verification error")
	}
	if len(env.runCommands) != 1 {
		t.Fatalf("run commands = %#v", env.runCommands)
	}
	env.assertNoMutation(t)
	if env.vars["GOROOT"] != "" || env.vars["DEVMAN_HOME"] != "" {
		t.Fatalf("user env changed before verification completed: %#v", env.vars)
	}
}

func TestSwitchVersionVerificationFailureDoesNotPersistVersionState(t *testing.T) {
	env := newFakeEnvironment()
	env.exeDir = `D:\apps\DevMan`
	env.files[`D:\production\go1.26\bin\go.exe`] = true
	env.runErr = errFakeRunFailed()
	reg := newFakeVersionRegistry([]ManagedVersion{
		{
			ID:          1,
			ToolKey:     "go",
			Version:     "1.25.0",
			InstallPath: `D:\production\go1.25`,
			IsDefault:   true,
			IsActive:    true,
		},
		{
			ID:          2,
			ToolKey:     "go",
			Version:     "1.26.0",
			InstallPath: `D:\production\go1.26`,
		},
	})

	_, err := NewService(reg, env).SwitchVersion(reg.versions[1])
	if err == nil {
		t.Fatal("expected verification error")
	}
	if len(reg.saved) != 0 {
		t.Fatalf("versions were saved before full success: %#v", reg.saved)
	}
	if !reg.versions[0].IsDefault || !reg.versions[0].IsActive {
		t.Fatalf("existing version state changed: default:%t active:%t", reg.versions[0].IsDefault, reg.versions[0].IsActive)
	}
	if reg.versions[1].IsDefault || reg.versions[1].IsActive {
		t.Fatalf("selected version state changed before full success: default:%t active:%t", reg.versions[1].IsDefault, reg.versions[1].IsActive)
	}
	env.assertNoMutation(t)
}

func TestSwitchVersionRejectsBadInstallPathBeforeMutation(t *testing.T) {
	tests := []struct {
		name        string
		installPath string
		wantError   string
	}{
		{name: "blank", installPath: "", wantError: "invalid install path: required"},
		{name: "whitespace", installPath: "   ", wantError: "invalid install path: required"},
		{name: "relative", installPath: `go1.26`, wantError: "invalid install path: must be absolute"},
		{name: "drive root", installPath: `D:\`, wantError: "invalid install path: must not be drive root"},
		{name: "filesystem root", installPath: `\`, wantError: "invalid install path: must be absolute"},
		{name: "quote", installPath: `D:\production\go"1.26`, wantError: "invalid install path: contains quote"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newFakeEnvironment()
			env.exeDir = `D:\apps\DevMan`

			_, err := NewService(nil, env).SwitchVersion(ManagedVersion{
				ToolKey:     "go",
				Version:     "1.26.0",
				InstallPath: tt.installPath,
			})
			if err == nil {
				t.Fatal("expected install path error")
			}
			if err.Error() != tt.wantError {
				t.Fatalf("error = %q, want %q", err, tt.wantError)
			}
			if len(env.runCommands) != 0 {
				t.Fatalf("run commands occurred before install path validation completed: %#v", env.runCommands)
			}
			env.assertNoMutation(t)
		})
	}
}

func TestSwitchVersionRejectsMissingExecutableBeforeMutation(t *testing.T) {
	env := newFakeEnvironment()
	env.exeDir = `D:\apps\DevMan`

	_, err := NewService(nil, env).SwitchVersion(ManagedVersion{
		ToolKey:     "go",
		Version:     "1.26.0",
		InstallPath: `D:\production\go1.26`,
	})
	if err == nil {
		t.Fatal("expected missing executable error")
	}
	if err.Error() != `expected executable not found: D:\production\go1.26\bin\go.exe` {
		t.Fatalf("error = %q", err)
	}
	env.assertNoMutation(t)
}

func TestSwitchVersionRejectsUnsupportedToolBeforeConflictDetection(t *testing.T) {
	env := newFakeEnvironment()
	env.paths["asdf"] = `C:\tools\asdf\asdf.exe`

	_, err := NewService(nil, env).SwitchVersion(ManagedVersion{ToolKey: "unknown", Version: "1.0.0"})
	if err == nil {
		t.Fatal("expected unsupported tool error")
	}
	if err.Error() != "unsupported tool: unknown" {
		t.Fatalf("error = %q", err)
	}
}

func TestSwitchVersionBlocksVersionManagerConflict(t *testing.T) {
	env := newFakeEnvironment()
	env.vars["NVM_HOME"] = `C:\Users\me\AppData\Roaming\nvm`

	_, err := NewService(nil, env).SwitchVersion(ManagedVersion{ToolKey: "node", Version: "22.11.0"})
	if err == nil {
		t.Fatal("expected conflict error")
	}
	if err.Error() != "node is managed by nvm; DevMan will not take over this tool" {
		t.Fatalf("error = %q", err)
	}
}

func TestSwitchVersionRejectsNonMutableEnvironment(t *testing.T) {
	env := readOnlyEnvironment{}

	_, err := NewService(nil, env).SwitchVersion(ManagedVersion{ToolKey: "go", Version: "1.26.0"})
	if err == nil {
		t.Fatal("expected mutation support error")
	}
	if err.Error() != "environment does not support mutation" {
		t.Fatalf("error = %q", err)
	}
}

type readOnlyEnvironment struct{}

func (readOnlyEnvironment) Getenv(string) string   { return "" }
func (readOnlyEnvironment) LookPath(string) string { return "" }
func (readOnlyEnvironment) DirExists(string) bool  { return false }
func (readOnlyEnvironment) FileExists(string) bool { return false }
