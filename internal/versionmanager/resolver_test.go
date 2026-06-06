package versionmanager

import "testing"

func TestGoResolverUsesGOROOTSibling(t *testing.T) {
	env := newFakeEnvironment()
	env.vars["GOROOT"] = `D:\production\go1.26`

	plan, err := ResolveInstallRoot(env, "go", "1.25.0")
	if err != nil {
		t.Fatalf("ResolveInstallRoot failed: %v", err)
	}
	if plan.TargetDir != `D:\production\go1.25.0` {
		t.Fatalf("TargetDir = %q", plan.TargetDir)
	}
	if plan.ResolverReason == "" {
		t.Fatal("expected resolver reason")
	}
}

func TestNodeResolverUsesNODEHOME(t *testing.T) {
	env := newFakeEnvironment()
	env.vars["NODE_HOME"] = `D:\production\nodejs`

	plan, err := ResolveInstallRoot(env, "node", "22.11.0")
	if err != nil {
		t.Fatalf("ResolveInstallRoot failed: %v", err)
	}
	if plan.TargetDir != `D:\production\node-v22.11.0` {
		t.Fatalf("TargetDir = %q", plan.TargetDir)
	}
}

func TestFlutterResolverUsesFlutterRootSibling(t *testing.T) {
	env := newFakeEnvironment()
	env.vars["FLUTTER_ROOT"] = `D:\production\flutter`

	plan, err := ResolveInstallRoot(env, "flutter", "3.24.5")
	if err != nil {
		t.Fatalf("ResolveInstallRoot failed: %v", err)
	}
	if plan.TargetDir != `D:\production\flutter-3.24.5` {
		t.Fatalf("TargetDir = %q", plan.TargetDir)
	}
}
