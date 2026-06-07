package versionmanager

import "testing"

func TestPreviewInstallBlocksVersionManagerOwnedTool(t *testing.T) {
	env := newFakeEnvironment()
	env.vars["NVM_HOME"] = `C:\Users\me\AppData\Roaming\nvm`

	_, err := NewService(nil, env).PreviewVersionInstall("node", "22.11.0")
	if err == nil {
		t.Fatal("expected conflict error")
	}
	if err.Error() != "node is managed by nvm; DevMan will not take over this tool" {
		t.Fatalf("error = %q", err)
	}
}

func TestPreviewInstallReturnsResolvedInstallPlan(t *testing.T) {
	env := newFakeEnvironment()
	env.vars["GOROOT"] = `D:\production\go1.26`
	env.dirs[`D:\production\go1.26`] = true

	plan, err := NewService(nil, env).PreviewVersionInstall("go", "1.25.0")
	if err != nil {
		t.Fatalf("PreviewVersionInstall failed: %v", err)
	}
	if plan.TargetDir != `D:\production\go1.25.0` {
		t.Fatalf("TargetDir = %q", plan.TargetDir)
	}
}
