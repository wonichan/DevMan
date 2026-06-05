package main

import (
	"devman/internal/models"
	"devman/internal/registry"
	"path/filepath"
	"testing"
)

func setupTestApp(t *testing.T) *App {
	t.Helper()

	registry.DbPathOverride = filepath.Join(t.TempDir(), "devman.db")
	reg, err := registry.Open()
	if err != nil {
		t.Fatalf("open registry failed: %v", err)
	}

	t.Cleanup(func() {
		err := reg.Close()
		registry.DbPathOverride = ""
		if err != nil {
			t.Fatalf("close registry failed: %v", err)
		}
	})

	return &App{reg: reg}
}

func TestSaveScanMetricSnapshotsRecordsPerEnvAndAggregate(t *testing.T) {
	app := setupTestApp(t)

	app.saveScanMetricSnapshots([]models.EnvSummary{
		{
			Env:       models.Env{Key: "nodejs"},
			TotalSize: 100,
		},
		{
			Env:       models.Env{Key: "go"},
			TotalSize: 200,
		},
		{
			Env:       models.Env{Key: "empty"},
			TotalSize: 0,
		},
	})

	nodeSnapshots, err := app.reg.ListMetricSnapshots("env_total_size", "nodejs", 10)
	if err != nil {
		t.Fatalf("list node snapshots failed: %v", err)
	}
	if len(nodeSnapshots) != 1 || nodeSnapshots[0].ValueBytes != 100 {
		t.Fatalf("expected nodejs snapshot value 100, got %#v", nodeSnapshots)
	}

	allSnapshots, err := app.reg.ListMetricSnapshots("env_total_size", "all", 10)
	if err != nil {
		t.Fatalf("list aggregate snapshots failed: %v", err)
	}
	if len(allSnapshots) != 1 || allSnapshots[0].ValueBytes != 300 {
		t.Fatalf("expected aggregate snapshot value 300, got %#v", allSnapshots)
	}

	emptySnapshots, err := app.reg.ListMetricSnapshots("env_total_size", "empty", 10)
	if err != nil {
		t.Fatalf("list empty snapshots failed: %v", err)
	}
	if len(emptySnapshots) != 0 {
		t.Fatalf("expected no zero-size snapshots, got %#v", emptySnapshots)
	}
}

func TestSaveCleanMetricSnapshotRecordsFreedBytes(t *testing.T) {
	app := setupTestApp(t)

	app.saveCleanMetricSnapshot(123)

	snapshots, err := app.reg.ListMetricSnapshots("clean_freed_bytes", "all", 10)
	if err != nil {
		t.Fatalf("list clean snapshots failed: %v", err)
	}
	if len(snapshots) != 1 || snapshots[0].ValueBytes != 123 {
		t.Fatalf("expected clean snapshot value 123, got %#v", snapshots)
	}
}

func TestSaveCleanMetricSnapshotSkipsZeroBytes(t *testing.T) {
	app := setupTestApp(t)

	app.saveCleanMetricSnapshot(0)

	snapshots, err := app.reg.ListMetricSnapshots("clean_freed_bytes", "all", 10)
	if err != nil {
		t.Fatalf("list clean snapshots failed: %v", err)
	}
	if len(snapshots) != 0 {
		t.Fatalf("expected no zero-byte clean snapshot, got %#v", snapshots)
	}
}

func TestManageAndUnmanageEnv(t *testing.T) {
	app := setupTestApp(t)

	env := &models.Env{Name: "Go", Key: "go", Category: models.CategoryRuntime}
	if err := app.reg.SaveEnv(env); err != nil {
		t.Fatalf("save env failed: %v", err)
	}

	managed, err := app.ManageEnv("go")
	if err != nil {
		t.Fatalf("manage env failed: %v", err)
	}
	if !managed.IsManaged {
		t.Fatal("expected managed env")
	}

	unmanaged, err := app.UnmanageEnv("go")
	if err != nil {
		t.Fatalf("unmanage env failed: %v", err)
	}
	if unmanaged.IsManaged {
		t.Fatal("expected unmanaged env")
	}

	history, err := app.reg.GetHistory(10)
	if err != nil {
		t.Fatalf("get history failed: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("expected 2 management history entries, got %#v", history)
	}
	if history[0].Action != "unmanage_env" || history[1].Action != "manage_env" {
		t.Fatalf("unexpected history order/actions: %#v", history)
	}
}

func TestManageEnvRejectsEmptyKey(t *testing.T) {
	app := setupTestApp(t)

	if _, err := app.ManageEnv(" "); err == nil {
		t.Fatal("expected empty key to fail")
	}
}
