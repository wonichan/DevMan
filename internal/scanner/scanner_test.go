package scanner

import (
	"devman/internal/models"
	"devman/internal/registry"
	"devman/internal/versionmanager"
	"errors"
	"path/filepath"
	"testing"
)

type fakeScanner struct {
	name      string
	instances []models.EnvInstance
	paths     []models.EnvPath
	err       error
}

func (s fakeScanner) Name() string {
	return s.name
}

func (s fakeScanner) Detect() ([]models.EnvInstance, []models.EnvPath, error) {
	if s.err != nil {
		return nil, nil, s.err
	}
	if s.instances != nil || s.paths != nil {
		return s.instances, s.paths, nil
	}
	return []models.EnvInstance{
			{
				Version:     "detected",
				InstallPath: filepath.Join("C:", "Tools", s.name),
				IsActive:    true,
			},
		},
		[]models.EnvPath{
			{
				Type:      models.PathInstall,
				Path:      filepath.Join("C:", "Tools", s.name),
				SizeBytes: 1,
				IsMovable: true,
			},
		},
		nil
}

func setupScannerRegistry(t *testing.T) (*registry.Registry, func()) {
	t.Helper()
	registry.DbPathOverride = filepath.Join(t.TempDir(), "devman.db")
	reg, err := registry.Open()
	if err != nil {
		t.Fatalf("open registry failed: %v", err)
	}
	return reg, func() {
		if err := reg.Close(); err != nil {
			t.Fatalf("close registry failed: %v", err)
		}
		registry.DbPathOverride = ""
	}
}

func TestScanAllReturnsPreservedManagedState(t *testing.T) {
	reg, cleanup := setupScannerRegistry(t)
	defer cleanup()

	env := &models.Env{Name: "CustomTool", Key: "customtool", Category: models.CategoryTool}
	if err := reg.SaveEnv(env); err != nil {
		t.Fatalf("save env failed: %v", err)
	}
	if _, err := reg.SetEnvManaged("customtool", true); err != nil {
		t.Fatalf("set env managed failed: %v", err)
	}

	engine := &Engine{reg: reg, scanners: []Scanner{fakeScanner{name: "customtool"}}}
	summaries, err := engine.ScanAll()
	if err != nil {
		t.Fatalf("scan all failed: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if !summaries[0].Env.IsManaged {
		t.Fatal("scan summary should preserve managed state")
	}
}

func TestScanAllSyncsDetectedVersionsIntoToolVersions(t *testing.T) {
	reg, cleanup := setupScannerRegistry(t)
	defer cleanup()

	installPath := filepath.Join("D:", "toolchains", "go")
	engine := &Engine{reg: reg, scanners: []Scanner{fakeScanner{
		name: "go",
		instances: []models.EnvInstance{{
			Version:     "go version go1.25.3 windows/amd64",
			InstallPath: installPath,
			IsDefault:   true,
			IsActive:    true,
		}},
		paths: []models.EnvPath{{
			Type:      models.PathInstall,
			Path:      installPath,
			SizeBytes: 1,
			IsMovable: true,
		}},
	}}}

	if _, err := engine.ScanAll(); err != nil {
		t.Fatalf("scan all failed: %v", err)
	}

	versions, err := reg.ListToolVersions("go")
	if err != nil {
		t.Fatalf("ListToolVersions failed: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("expected synced tool version, got %#v", versions)
	}
	got := versions[0]
	if got.Version != "go version go1.25.3 windows/amd64" || got.InstallPath != installPath {
		t.Fatalf("unexpected synced version: %#v", got)
	}
	if got.Source != versionmanager.SourceExternal || !got.IsDefault || !got.IsActive {
		t.Fatalf("unexpected synced flags: %#v", got)
	}
	if got.BinPath != filepath.Join(installPath, "bin", "go.exe") {
		t.Fatalf("expected synced bin path, got %#v", got)
	}
}

func TestScanAllPreservesExistingManagedToolVersionStateDuringSync(t *testing.T) {
	reg, cleanup := setupScannerRegistry(t)
	defer cleanup()

	installPath := filepath.Join("D:", "toolchains", "go")
	existing := &versionmanager.ManagedVersion{
		ToolKey:      "go",
		Version:      "1.25.0",
		InstallPath:  installPath,
		BinPath:      filepath.Join(installPath, "bin", "go.exe"),
		Source:       versionmanager.SourceDevMan,
		IsDefault:    true,
		IsActive:     true,
		CanDelete:    true,
		DeletePolicy: versionmanager.DeletePolicyDirect,
	}
	if err := reg.SaveToolVersion(existing); err != nil {
		t.Fatalf("SaveToolVersion failed: %v", err)
	}

	engine := &Engine{reg: reg, scanners: []Scanner{fakeScanner{
		name: "go",
		instances: []models.EnvInstance{{
			Version:     "go version go1.25.3 windows/amd64",
			InstallPath: installPath,
			IsDefault:   false,
			IsActive:    false,
		}},
		paths: []models.EnvPath{{
			Type:      models.PathInstall,
			Path:      installPath,
			SizeBytes: 1,
			IsMovable: true,
		}},
	}}}

	if _, err := engine.ScanAll(); err != nil {
		t.Fatalf("scan all failed: %v", err)
	}

	versions, err := reg.ListToolVersions("go")
	if err != nil {
		t.Fatalf("ListToolVersions failed: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("expected 1 synced version, got %#v", versions)
	}
	got := versions[0]
	if got.Source != versionmanager.SourceDevMan || !got.IsDefault || !got.IsActive || !got.CanDelete || got.DeletePolicy != versionmanager.DeletePolicyDirect {
		t.Fatalf("expected managed state preserved, got %#v", got)
	}
	if got.Version != "go version go1.25.3 windows/amd64" {
		t.Fatalf("expected detected version refresh, got %#v", got)
	}
}

func TestScanAllRemovesStaleUnmanagedEnvWhenScannerReportsNone(t *testing.T) {
	reg, cleanup := setupScannerRegistry(t)
	defer cleanup()

	stale := &models.Env{Name: "Flutter", Key: "flutter", Category: models.CategorySDK}
	if err := reg.SaveEnv(stale); err != nil {
		t.Fatalf("save stale env failed: %v", err)
	}
	if err := reg.SaveInstance(&models.EnvInstance{
		EnvID:       stale.ID,
		Version:     "3.19.0",
		InstallPath: filepath.Join("C:", "Tools", "flutter"),
		IsActive:    true,
		Source:      "system",
	}); err != nil {
		t.Fatalf("save stale instance failed: %v", err)
	}
	if err := reg.SavePath(&models.EnvPath{
		EnvID:     stale.ID,
		Type:      models.PathInstall,
		Path:      filepath.Join("C:", "Tools", "flutter"),
		SizeBytes: 1,
		IsMovable: true,
	}); err != nil {
		t.Fatalf("save stale path failed: %v", err)
	}

	engine := &Engine{reg: reg, scanners: []Scanner{fakeScanner{
		name:      "flutter",
		instances: []models.EnvInstance{},
		paths:     []models.EnvPath{},
	}}}

	if _, err := engine.ScanAll(); err != nil {
		t.Fatalf("scan all failed: %v", err)
	}

	got, err := reg.GetEnvByKey("flutter")
	if err != nil {
		t.Fatalf("get env failed: %v", err)
	}
	if got != nil {
		t.Fatalf("expected stale unmanaged flutter env to be removed, got %#v", got)
	}

	instances, err := reg.ListInstances(stale.ID)
	if err != nil {
		t.Fatalf("list instances failed: %v", err)
	}
	if len(instances) != 0 {
		t.Fatalf("expected stale instances to be removed, got %#v", instances)
	}
	paths, err := reg.ListPaths(stale.ID)
	if err != nil {
		t.Fatalf("list paths failed: %v", err)
	}
	if len(paths) != 0 {
		t.Fatalf("expected stale paths to be removed, got %#v", paths)
	}
}

func TestScanAllPreservesStaleManagedEnvWhenScannerReportsNone(t *testing.T) {
	reg, cleanup := setupScannerRegistry(t)
	defer cleanup()

	managed := &models.Env{Name: "Flutter", Key: "flutter", Category: models.CategorySDK}
	if err := reg.SaveEnv(managed); err != nil {
		t.Fatalf("save managed env failed: %v", err)
	}
	if _, err := reg.SetEnvManaged("flutter", true); err != nil {
		t.Fatalf("set env managed failed: %v", err)
	}

	engine := &Engine{reg: reg, scanners: []Scanner{fakeScanner{
		name:      "flutter",
		instances: []models.EnvInstance{},
		paths:     []models.EnvPath{},
	}}}

	if _, err := engine.ScanAll(); err != nil {
		t.Fatalf("scan all failed: %v", err)
	}

	got, err := reg.GetEnvByKey("flutter")
	if err != nil {
		t.Fatalf("get env failed: %v", err)
	}
	if got == nil {
		t.Fatal("expected managed flutter env to be preserved when scanner detects none")
	}
	if !got.IsManaged {
		t.Fatal("managed flutter env should remain managed after scan")
	}
}

func TestScanAllDoesNotTouchEnvsOutsideScannerSetWhenScannerReportsNone(t *testing.T) {
	reg, cleanup := setupScannerRegistry(t)
	defer cleanup()

	unrelated := &models.Env{Name: "Node.js", Key: "nodejs", Category: models.CategoryRuntime}
	if err := reg.SaveEnv(unrelated); err != nil {
		t.Fatalf("save unrelated env failed: %v", err)
	}

	engine := &Engine{reg: reg, scanners: []Scanner{fakeScanner{
		name:      "flutter",
		instances: []models.EnvInstance{},
		paths:     []models.EnvPath{},
	}}}

	if _, err := engine.ScanAll(); err != nil {
		t.Fatalf("scan all failed: %v", err)
	}

	got, err := reg.GetEnvByKey("nodejs")
	if err != nil {
		t.Fatalf("get env failed: %v", err)
	}
	if got == nil {
		t.Fatal("unrelated env outside scanner set must not be removed")
	}
}

func TestScanAllPreservesEnvWhenScannerFails(t *testing.T) {
	reg, cleanup := setupScannerRegistry(t)
	defer cleanup()

	existing := &models.Env{Name: "Flutter", Key: "flutter", Category: models.CategorySDK}
	if err := reg.SaveEnv(existing); err != nil {
		t.Fatalf("save existing env failed: %v", err)
	}

	engine := &Engine{reg: reg, scanners: []Scanner{fakeScanner{
		name: "flutter",
		err:  errors.New("scanner failed"),
	}}}

	if _, err := engine.ScanAll(); err != nil {
		t.Fatalf("scan all failed: %v", err)
	}

	got, err := reg.GetEnvByKey("flutter")
	if err != nil {
		t.Fatalf("get env failed: %v", err)
	}
	if got == nil {
		t.Fatal("scanner failure must not delete an existing env")
	}
}

func TestSyncedToolKeyMapsSupportedScannerKeys(t *testing.T) {
	tests := []struct {
		envKey   string
		wantKey  string
		wantOkay bool
	}{
		{envKey: "go", wantKey: "go", wantOkay: true},
		{envKey: "nodejs", wantKey: "node", wantOkay: true},
		{envKey: "bun", wantKey: "bun", wantOkay: true},
		{envKey: "flutter", wantKey: "flutter", wantOkay: true},
		{envKey: "python", wantKey: "", wantOkay: false},
	}

	for _, tt := range tests {
		gotKey, gotOkay := syncedToolKey(tt.envKey)
		if gotKey != tt.wantKey || gotOkay != tt.wantOkay {
			t.Fatalf("syncedToolKey(%q) = (%q, %t), want (%q, %t)", tt.envKey, gotKey, gotOkay, tt.wantKey, tt.wantOkay)
		}
	}
}
