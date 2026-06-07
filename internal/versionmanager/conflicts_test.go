package versionmanager

import "testing"

func TestNodeConflictDetectsNVMHome(t *testing.T) {
	env := newFakeEnvironment()
	env.vars["NVM_HOME"] = `C:\Users\me\AppData\Roaming\nvm`

	conflict := DetectVersionManagerConflict(env, "node")
	if conflict == nil || !conflict.Detected || conflict.Manager != "nvm" {
		t.Fatalf("expected nvm conflict, got %#v", conflict)
	}
	if conflict.ToolKey != "node" {
		t.Fatalf("ToolKey = %q, want node", conflict.ToolKey)
	}
	if conflict.Evidence != `NVM_HOME=C:\Users\me\AppData\Roaming\nvm` {
		t.Fatalf("Evidence = %q", conflict.Evidence)
	}
}

func TestNodeConflictDetectsFNMOnPath(t *testing.T) {
	env := newFakeEnvironment()
	env.paths["fnm"] = `C:\tools\fnm\fnm.exe`

	conflict := DetectVersionManagerConflict(env, "node")
	if conflict == nil || !conflict.Detected || conflict.Manager != "fnm" {
		t.Fatalf("expected fnm conflict, got %#v", conflict)
	}
	if conflict.Evidence != `fnm at C:\tools\fnm\fnm.exe` {
		t.Fatalf("Evidence = %q", conflict.Evidence)
	}
}

func TestFlutterConflictDetectsFVMOnPath(t *testing.T) {
	env := newFakeEnvironment()
	env.paths["fvm"] = `C:\tools\fvm\fvm.exe`

	conflict := DetectVersionManagerConflict(env, "flutter")
	if conflict == nil || !conflict.Detected || conflict.Manager != "fvm" {
		t.Fatalf("expected fvm conflict, got %#v", conflict)
	}
}

func TestGoConflictDetectsGVMOnPath(t *testing.T) {
	env := newFakeEnvironment()
	env.paths["gvm"] = `C:\tools\gvm\gvm.exe`

	conflict := DetectVersionManagerConflict(env, "go")
	if conflict == nil || !conflict.Detected || conflict.Manager != "gvm" {
		t.Fatalf("expected gvm conflict, got %#v", conflict)
	}
}

func TestCommonConflictDetectsASDFOnPath(t *testing.T) {
	for _, toolKey := range []string{"go", "node", "bun", "flutter", "unknown"} {
		t.Run(toolKey, func(t *testing.T) {
			env := newFakeEnvironment()
			env.paths["asdf"] = `C:\tools\asdf\asdf.exe`

			conflict := DetectVersionManagerConflict(env, toolKey)
			if conflict == nil || !conflict.Detected || conflict.Manager != "asdf" {
				t.Fatalf("expected asdf conflict, got %#v", conflict)
			}
			if conflict.ToolKey != toolKey {
				t.Fatalf("ToolKey = %q, want %q", conflict.ToolKey, toolKey)
			}
		})
	}
}

func TestNoVersionManagerConflictReturnsNil(t *testing.T) {
	conflict := DetectVersionManagerConflict(newFakeEnvironment(), "node")
	if conflict != nil {
		t.Fatalf("expected no conflict, got %#v", conflict)
	}
}
