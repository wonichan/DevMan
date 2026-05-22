package migrator

import (
	"devman/internal/models"
	"devman/internal/registry"
	"os"
	"path/filepath"
	"testing"
)

func setupTestMigrator(t *testing.T) (*Engine, *registry.Registry, func()) {
	dbPath := filepath.Join(t.TempDir(), "devman.db")
	registry.DbPathOverride = dbPath

	reg, err := registry.Open()
	if err != nil {
		t.Fatalf("open registry failed: %v", err)
	}

	engine := New(reg)
	cleanup := func() {
		reg.Close()
		registry.DbPathOverride = ""
	}
	return engine, reg, cleanup
}

func TestCopyDir(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir() + "_copy"

	// Create source files
	f1 := filepath.Join(srcDir, "file1.txt")
	os.WriteFile(f1, []byte("hello"), 0644)
	subDir := filepath.Join(srcDir, "subdir")
	os.MkdirAll(subDir, 0755)
	f2 := filepath.Join(subDir, "file2.txt")
	os.WriteFile(f2, []byte("world"), 0644)

	e := &Engine{}
	size, err := e.copyDir(srcDir, dstDir)
	if err != nil {
		t.Fatalf("copy dir failed: %v", err)
	}
	if size != 10 {
		t.Errorf("expected 10 bytes, got %d", size)
	}

	// Verify destination
	if _, err := os.Stat(filepath.Join(dstDir, "file1.txt")); err != nil {
		t.Error("file1 should exist in destination")
	}
	if _, err := os.Stat(filepath.Join(dstDir, "subdir", "file2.txt")); err != nil {
		t.Error("file2 should exist in destination subdir")
	}
}

func TestVerifyCopy(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir() + "_verify"

	os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(srcDir, "b.txt"), []byte("bb"), 0644)

	e := &Engine{}
	_, _ = e.copyDir(srcDir, dstDir)

	if !e.verifyCopy(srcDir, dstDir) {
		t.Error("verify should pass for identical copies")
	}

	// Modify destination
	os.WriteFile(filepath.Join(dstDir, "a.txt"), []byte("changed"), 0644)
	if e.verifyCopy(srcDir, dstDir) {
		t.Error("verify should fail for modified destination")
	}
}

func TestPreCheck(t *testing.T) {
	engine, reg, cleanup := setupTestMigrator(t)
	defer cleanup()

	// Create a mock env
	env := &models.Env{Name: "Test", Key: "test", Category: models.CategoryRuntime}
	reg.SaveEnv(env)

	// Save a path for this env
	p := &models.EnvPath{
		EnvID:     env.ID,
		Type:      models.PathInstall,
		Path:      t.TempDir(),
		SizeBytes: 100,
		IsMovable: true,
	}
	reg.SavePath(p)

	// Pre-check should pass with valid target
	targetDir := t.TempDir() + "_target"
	if err := engine.preCheck(env.ID, targetDir); err != nil {
		t.Errorf("pre-check should pass: %v", err)
	}

	// Pre-check should fail with invalid target (file instead of dir)
	tmpFile := filepath.Join(t.TempDir(), "file.txt")
	os.WriteFile(tmpFile, []byte("x"), 0644)
	if err := engine.preCheck(env.ID, tmpFile); err == nil {
		t.Error("pre-check should fail when target is a file")
	}
}

func TestSnapshot(t *testing.T) {
	engine, reg, cleanup := setupTestMigrator(t)
	defer cleanup()

	env := &models.Env{Name: "SnapTest", Key: "snaptest"}
	reg.SaveEnv(env)

	snap, err := engine.createSnapshot(env.ID)
	if err != nil {
		t.Fatalf("create snapshot failed: %v", err)
	}
	if snap == nil {
		t.Fatal("snapshot should not be nil")
	}
	if snap.ID == 0 {
		t.Error("snapshot ID should be assigned")
	}

	// Verify we can fetch it
	fetched, err := reg.GetSnapshot(snap.ID)
	if err != nil {
		t.Fatalf("get snapshot failed: %v", err)
	}
	if fetched == nil {
		t.Fatal("snapshot should exist")
	}
}
