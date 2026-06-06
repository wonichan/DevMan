package versionmanager

import (
	"strings"
	"testing"
)

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

func TestResolverEnvVarWinsWhenPathAlsoPresent(t *testing.T) {
	env := newFakeEnvironment()
	env.vars["GOROOT"] = `D:\env\go1.26`
	env.paths["go"] = `D:\path\go1.26\bin\go.exe`

	plan, err := ResolveInstallRoot(env, "go", "1.25.0")
	if err != nil {
		t.Fatalf("ResolveInstallRoot failed: %v", err)
	}
	if plan.TargetDir != `D:\env\go1.25.0` {
		t.Fatalf("TargetDir = %q", plan.TargetDir)
	}
}

func TestResolverTrimsAndCleansEnvRoot(t *testing.T) {
	env := newFakeEnvironment()
	env.vars["GOROOT"] = "  D:\\production\\go1.26\\  "

	plan, err := ResolveInstallRoot(env, "go", "1.25.0")
	if err != nil {
		t.Fatalf("ResolveInstallRoot failed: %v", err)
	}
	if plan.TargetDir != `D:\production\go1.25.0` {
		t.Fatalf("TargetDir = %q", plan.TargetDir)
	}
}

func TestGoResolverUsesPathExecutableBinParent(t *testing.T) {
	env := newFakeEnvironment()
	env.paths["go"] = `D:\production\go1.26\bin\go.exe`

	plan, err := ResolveInstallRoot(env, "go", "1.25.0")
	if err != nil {
		t.Fatalf("ResolveInstallRoot failed: %v", err)
	}
	if plan.TargetDir != `D:\production\go1.25.0` {
		t.Fatalf("TargetDir = %q", plan.TargetDir)
	}
}

func TestFlutterResolverUsesPathExecutableBinParent(t *testing.T) {
	env := newFakeEnvironment()
	env.paths["flutter"] = `D:\tools\flutter\bin\flutter.bat`

	plan, err := ResolveInstallRoot(env, "flutter", "3.24.5")
	if err != nil {
		t.Fatalf("ResolveInstallRoot failed: %v", err)
	}
	if plan.TargetDir != `D:\tools\flutter-3.24.5` {
		t.Fatalf("TargetDir = %q", plan.TargetDir)
	}
}

func TestBunResolverUsesPathExecutableCmdParent(t *testing.T) {
	env := newFakeEnvironment()
	env.paths["bun"] = `D:\tools\bun\cmd\bun.exe`

	plan, err := ResolveInstallRoot(env, "bun", "1.2.0")
	if err != nil {
		t.Fatalf("ResolveInstallRoot failed: %v", err)
	}
	if plan.TargetDir != `D:\tools\bun-v1.2.0` {
		t.Fatalf("TargetDir = %q", plan.TargetDir)
	}
}

func TestResolverRejectsBlankVersion(t *testing.T) {
	env := newFakeEnvironment()
	env.vars["GOROOT"] = `D:\production\go1.26`

	if _, err := ResolveInstallRoot(env, "go", "   "); err == nil {
		t.Fatal("expected error")
	}
}

func TestResolverRejectsUnsupportedTool(t *testing.T) {
	env := newFakeEnvironment()

	if _, err := ResolveInstallRoot(env, "ruby", "3.3.0"); err == nil {
		t.Fatal("expected error")
	}
}

func TestResolverWhitespaceEnvVarFallsThroughToPath(t *testing.T) {
	env := newFakeEnvironment()
	env.vars["GOROOT"] = "   "
	env.paths["go"] = `D:\path\go1.26\bin\go.exe`

	plan, err := ResolveInstallRoot(env, "go", "1.25.0")
	if err != nil {
		t.Fatalf("ResolveInstallRoot failed: %v", err)
	}
	if plan.TargetDir != `D:\path\go1.25.0` {
		t.Fatalf("TargetDir = %q", plan.TargetDir)
	}
}

func TestResolverSetsWillOverwriteForExistingTargetDir(t *testing.T) {
	env := newFakeEnvironment()
	env.vars["GOROOT"] = `D:\production\go1.26`
	env.dirs[`D:\production\go1.25.0`] = true

	plan, err := ResolveInstallRoot(env, "go", "1.25.0")
	if err != nil {
		t.Fatalf("ResolveInstallRoot failed: %v", err)
	}
	if !plan.WillOverwrite {
		t.Fatal("expected WillOverwrite")
	}
}

func TestResolverTrimsVersionForPlanAndTargetDir(t *testing.T) {
	env := newFakeEnvironment()
	env.vars["GOROOT"] = `D:\production\go1.26`

	plan, err := ResolveInstallRoot(env, "go", "  1.25.0  ")
	if err != nil {
		t.Fatalf("ResolveInstallRoot failed: %v", err)
	}
	if plan.Version != "1.25.0" {
		t.Fatalf("Version = %q", plan.Version)
	}
	if plan.TargetDir != `D:\production\go1.25.0` {
		t.Fatalf("TargetDir = %q", plan.TargetDir)
	}
}

func TestResolverRelativeEnvRootFallsThroughToPath(t *testing.T) {
	env := newFakeEnvironment()
	env.vars["GOROOT"] = "go1.26"
	env.paths["go"] = `D:\path\go1.26\bin\go.exe`

	plan, err := ResolveInstallRoot(env, "go", "1.25.0")
	if err != nil {
		t.Fatalf("ResolveInstallRoot failed: %v", err)
	}
	if plan.TargetDir != `D:\path\go1.25.0` {
		t.Fatalf("TargetDir = %q", plan.TargetDir)
	}
}

func TestResolverDotDotEnvRootFallsThroughToPath(t *testing.T) {
	env := newFakeEnvironment()
	env.vars["GOROOT"] = ".."
	env.paths["go"] = `D:\path\go1.26\bin\go.exe`

	plan, err := ResolveInstallRoot(env, "go", "1.25.0")
	if err != nil {
		t.Fatalf("ResolveInstallRoot failed: %v", err)
	}
	if plan.TargetDir != `D:\path\go1.25.0` {
		t.Fatalf("TargetDir = %q", plan.TargetDir)
	}
}

func TestResolverDriveRelativeEnvRootFallsThroughToPath(t *testing.T) {
	env := newFakeEnvironment()
	env.vars["GOROOT"] = "D:"
	env.paths["go"] = `D:\path\go1.26\bin\go.exe`

	plan, err := ResolveInstallRoot(env, "go", "1.25.0")
	if err != nil {
		t.Fatalf("ResolveInstallRoot failed: %v", err)
	}
	if plan.TargetDir != `D:\path\go1.25.0` {
		t.Fatalf("TargetDir = %q", plan.TargetDir)
	}
}

func TestResolverInvalidEnvRootWithoutPathReturnsCannotInfer(t *testing.T) {
	env := newFakeEnvironment()
	env.vars["GOROOT"] = "go1.26"

	_, err := ResolveInstallRoot(env, "go", "1.25.0")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "cannot infer install root") {
		t.Fatalf("error = %q", err)
	}
}

func TestResolverRejectsPathLikeVersions(t *testing.T) {
	env := newFakeEnvironment()
	env.vars["GOROOT"] = `D:\production\go1.26`

	for _, version := range []string{"../1.25.0", "1.25/evil", `1.25\evil`, "C:1.25"} {
		t.Run(version, func(t *testing.T) {
			if _, err := ResolveInstallRoot(env, "go", version); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestResolverAllowsLettersDigitsDotsUnderscoresAndDashesInVersion(t *testing.T) {
	env := newFakeEnvironment()
	env.vars["NODE_HOME"] = `D:\production\nodejs`

	plan, err := ResolveInstallRoot(env, "node", "22.11.0-rc_1")
	if err != nil {
		t.Fatalf("ResolveInstallRoot failed: %v", err)
	}
	if plan.TargetDir != `D:\production\node-v22.11.0-rc_1` {
		t.Fatalf("TargetDir = %q", plan.TargetDir)
	}
}
