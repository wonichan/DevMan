package scanner

import (
	"devman/internal/models"
	"devman/internal/registry"
	"path/filepath"
	"testing"
)

type fakeScanner struct {
	name string
}

func (s fakeScanner) Name() string {
	return s.name
}

func (s fakeScanner) Detect() ([]models.EnvInstance, []models.EnvPath, error) {
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
