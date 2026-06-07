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

	service := NewService(nil, env)
	service.versionProvider = fakeOfficialCatalog("go", "1.25.0", "https://go.dev/dl/go1.25.0.windows-amd64.zip")

	plan, err := service.PreviewVersionInstall("go", "1.25.0")
	if err != nil {
		t.Fatalf("PreviewVersionInstall failed: %v", err)
	}
	if plan.TargetDir != `D:\production\go1.25.0` {
		t.Fatalf("TargetDir = %q", plan.TargetDir)
	}
}

func TestPreviewInstallPopulatesOfficialDownloadMetadata(t *testing.T) {
	env := newFakeEnvironment()
	env.vars["GOROOT"] = `D:\production\go1.26`
	env.dirs[`D:\production\go1.26`] = true
	service := NewService(nil, env)
	service.versionProvider = fakeOfficialVersionProvider{
		catalog: &ToolVersionCatalog{
			ToolKey: "go",
			Versions: []AvailableVersion{{
				Version:     "1.25.0",
				DownloadURL: "https://go.dev/dl/go1.25.0.windows-amd64.zip",
			}},
		},
	}

	plan, err := service.PreviewVersionInstall("go", "go1.25.0")
	if err != nil {
		t.Fatalf("PreviewVersionInstall failed: %v", err)
	}
	if plan.Version != "1.25.0" {
		t.Fatalf("Version = %q, want canonical version", plan.Version)
	}
	if plan.DownloadURL != "https://go.dev/dl/go1.25.0.windows-amd64.zip" {
		t.Fatalf("DownloadURL = %q", plan.DownloadURL)
	}
	if plan.ArchiveName != "go1.25.0.windows-amd64.zip" {
		t.Fatalf("ArchiveName = %q", plan.ArchiveName)
	}
}

func TestInstallVersionUsesOfficialDownloadMetadata(t *testing.T) {
	env := newFakeEnvironment()
	env.vars["NODE_HOME"] = `D:\production\node-v22.10.0`
	env.dirs[`D:\production\node-v22.10.0`] = true
	reg := newFakeVersionRegistry(nil)
	downloader := &fakeDownloader{}
	service := NewService(reg, env)
	service.downloader = downloader
	service.versionProvider = fakeOfficialVersionProvider{
		catalog: &ToolVersionCatalog{
			ToolKey: "node",
			Versions: []AvailableVersion{{
				Version:     "22.11.0",
				DownloadURL: "https://nodejs.org/dist/v22.11.0/node-v22.11.0-win-x64.zip",
			}},
		},
	}

	_, err := service.InstallVersion("node", "v22.11.0", `D:\production\node-v22.11.0`)
	if err != nil {
		t.Fatalf("InstallVersion returned error: %v", err)
	}
	if len(downloader.plans) != 1 {
		t.Fatalf("downloader plans = %d, want 1", len(downloader.plans))
	}
	plan := downloader.plans[0]
	if plan.Version != "v22.11.0" {
		t.Fatalf("Version = %q, want requested canonical node version", plan.Version)
	}
	if plan.DownloadURL != "https://nodejs.org/dist/v22.11.0/node-v22.11.0-win-x64.zip" {
		t.Fatalf("DownloadURL = %q", plan.DownloadURL)
	}
	if plan.ArchiveName != "node-v22.11.0-win-x64.zip" {
		t.Fatalf("ArchiveName = %q", plan.ArchiveName)
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

func fakeOfficialCatalog(toolKey string, version string, downloadURL string) fakeOfficialVersionProvider {
	return fakeOfficialVersionProvider{
		catalog: &ToolVersionCatalog{
			ToolKey: toolKey,
			Versions: []AvailableVersion{{
				Version:     version,
				DownloadURL: downloadURL,
			}},
			FetchedAt: time.Now(),
			SourceURL: "fake://" + toolKey,
		},
	}
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
