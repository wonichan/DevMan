package versionmanager

import (
	"fmt"
	"testing"
	"time"
)

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

func TestPreviewInstallRejectsUnsupportedToolBeforeConflictDetection(t *testing.T) {
	env := newFakeEnvironment()
	env.paths["asdf"] = `C:\tools\asdf\asdf.exe`

	_, err := NewService(nil, env).PreviewVersionInstall("unknown", "1.0.0")
	if err == nil {
		t.Fatal("expected unsupported tool error")
	}
	if err.Error() != "unsupported tool: unknown" {
		t.Fatalf("error = %q", err)
	}
}

func TestListToolVersionsReturnsNonNilEmptyLocalVersions(t *testing.T) {
	states, err := NewService(nil, newFakeEnvironment()).ListToolVersions()
	if err != nil {
		t.Fatalf("ListToolVersions failed: %v", err)
	}
	if len(states) == 0 {
		t.Fatal("expected supported tool states")
	}
	for _, state := range states {
		if state.LocalVersions == nil {
			t.Fatalf("%s LocalVersions is nil", state.ToolKey)
		}
		if len(state.LocalVersions) != 0 {
			t.Fatalf("%s LocalVersions length = %d", state.ToolKey, len(state.LocalVersions))
		}
	}
}

func TestFetchOfficialVersionsUsesInjectedProvider(t *testing.T) {
	service := NewService(nil, newFakeEnvironment())
	service.versionProvider = fakeOfficialVersionProvider{
		catalog: &ToolVersionCatalog{
			ToolKey:   "go",
			Versions:  []AvailableVersion{{Version: "1.25.0"}},
			FetchedAt: time.Now(),
			SourceURL: "fake://go",
		},
	}

	catalog, err := service.FetchOfficialVersions("go")
	if err != nil {
		t.Fatal(err)
	}
	if catalog.ToolKey != "go" {
		t.Fatalf("ToolKey = %q", catalog.ToolKey)
	}
	if len(catalog.Versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(catalog.Versions))
	}
}

type fakeOfficialVersionProvider struct {
	catalog *ToolVersionCatalog
	err     error
}

func (p fakeOfficialVersionProvider) Fetch(toolKey string) (*ToolVersionCatalog, error) {
	if p.err != nil {
		return nil, p.err
	}
	if p.catalog == nil {
		return nil, fmt.Errorf("no fake catalog for %s", toolKey)
	}
	return p.catalog, nil
}
