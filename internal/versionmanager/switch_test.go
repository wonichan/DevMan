package versionmanager

import "testing"

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

func TestSwitchVersionRejectsBadInstallPathBeforeMutation(t *testing.T) {
	env := newFakeEnvironment()
	env.exeDir = `D:\apps\DevMan`

	_, err := NewService(nil, env).SwitchVersion(ManagedVersion{
		ToolKey:     "go",
		Version:     "1.26.0",
		InstallPath: `go1.26`,
	})
	if err == nil {
		t.Fatal("expected install path error")
	}
	if err.Error() != "invalid install path: must be absolute" {
		t.Fatalf("error = %q", err)
	}
	env.assertNoMutation(t)
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
