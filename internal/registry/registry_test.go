package registry

import (
	"devman/internal/models"
	"path/filepath"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) (*Registry, func()) {
	dbPath := filepath.Join(t.TempDir(), "devman.db")
	DbPathOverride = dbPath

	reg, err := Open()
	if err != nil {
		t.Fatalf("open registry failed: %v", err)
	}

	cleanup := func() {
		reg.Close()
		DbPathOverride = ""
	}
	return reg, cleanup
}

func TestOpenAndMigrate(t *testing.T) {
	_, cleanup := setupTestDB(t)
	defer cleanup()
}

func TestSaveAndGetEnv(t *testing.T) {
	reg, cleanup := setupTestDB(t)
	defer cleanup()

	env := &models.Env{
		Name:      "Node.js",
		Key:       "nodejs",
		Category:  models.CategoryRuntime,
		Icon:      "⚡",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := reg.SaveEnv(env); err != nil {
		t.Fatalf("save env failed: %v", err)
	}

	if env.ID == 0 {
		t.Error("env ID should be assigned after insert")
	}

	fetched, err := reg.GetEnvByKey("nodejs")
	if err != nil {
		t.Fatalf("get env failed: %v", err)
	}
	if fetched == nil {
		t.Fatal("env should exist")
	}
	if fetched.Name != "Node.js" {
		t.Errorf("expected name 'Node.js', got %s", fetched.Name)
	}

	list, err := reg.ListEnvs()
	if err != nil {
		t.Fatalf("list envs failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 env, got %d", len(list))
	}
}

func TestSaveEnvReloadsIDAfterConflict(t *testing.T) {
	reg, cleanup := setupTestDB(t)
	defer cleanup()

	first := &models.Env{Name: "Node.js", Key: "nodejs", Category: models.CategoryRuntime}
	if err := reg.SaveEnv(first); err != nil {
		t.Fatalf("save first env failed: %v", err)
	}
	if first.ID == 0 {
		t.Fatal("first env ID should be assigned")
	}

	second := &models.Env{Name: "Node.js Updated", Key: "nodejs", Category: models.CategoryRuntime}
	if err := reg.SaveEnv(second); err != nil {
		t.Fatalf("save conflicting env failed: %v", err)
	}
	if second.ID != first.ID {
		t.Fatalf("expected conflicting save to reload id %d, got %d", first.ID, second.ID)
	}
}

func TestInstancesAndPaths(t *testing.T) {
	reg, cleanup := setupTestDB(t)
	defer cleanup()

	env := &models.Env{Name: "Go", Key: "go", Category: models.CategoryRuntime}
	if err := reg.SaveEnv(env); err != nil {
		t.Fatalf("save env failed: %v", err)
	}

	inst := &models.EnvInstance{
		EnvID:       env.ID,
		Version:     "1.22.0",
		InstallPath: "/usr/local/go",
		IsDefault:   true,
		IsActive:    true,
	}
	if err := reg.SaveInstance(inst); err != nil {
		t.Fatalf("save instance failed: %v", err)
	}
	if inst.ID == 0 {
		t.Error("instance ID should be assigned")
	}

	instances, err := reg.ListInstances(env.ID)
	if err != nil {
		t.Fatalf("list instances failed: %v", err)
	}
	if len(instances) != 1 {
		t.Errorf("expected 1 instance, got %d", len(instances))
	}

	p := &models.EnvPath{
		EnvID:     env.ID,
		Type:      models.PathInstall,
		Path:      "/usr/local/go",
		SizeBytes: 500 * 1024 * 1024,
		IsMovable: true,
	}
	if err := reg.SavePath(p); err != nil {
		t.Fatalf("save path failed: %v", err)
	}

	paths, err := reg.ListPaths(env.ID)
	if err != nil {
		t.Fatalf("list paths failed: %v", err)
	}
	if len(paths) != 1 {
		t.Errorf("expected 1 path, got %d", len(paths))
	}
}

func TestHistory(t *testing.T) {
	reg, cleanup := setupTestDB(t)
	defer cleanup()

	entry := &models.HistoryEntry{
		Action:    "migrate",
		TargetEnv: "nodejs",
		Success:   true,
		CreatedAt: time.Now(),
	}
	if err := reg.SaveHistory(entry); err != nil {
		t.Fatalf("save history failed: %v", err)
	}

	history, err := reg.GetHistory(10)
	if err != nil {
		t.Fatalf("get history failed: %v", err)
	}
	if len(history) != 1 {
		t.Errorf("expected 1 history entry, got %d", len(history))
	}
	if history[0].Action != "migrate" {
		t.Errorf("expected action 'migrate', got %s", history[0].Action)
	}
}

func TestSnapshot(t *testing.T) {
	reg, cleanup := setupTestDB(t)
	defer cleanup()

	snap := &models.Snapshot{
		Name:      "test-snap",
		DataJSON:  `{"test":"data"}`,
		CreatedAt: time.Now(),
	}
	if err := reg.SaveSnapshot(snap); err != nil {
		t.Fatalf("save snapshot failed: %v", err)
	}
	if snap.ID == 0 {
		t.Error("snapshot ID should be assigned")
	}

	fetched, err := reg.GetSnapshot(snap.ID)
	if err != nil {
		t.Fatalf("get snapshot failed: %v", err)
	}
	if fetched == nil {
		t.Fatal("snapshot should exist")
	}
	if fetched.Name != "test-snap" {
		t.Errorf("expected name 'test-snap', got %s", fetched.Name)
	}
}

func TestImportSnapshotDataRestoresInstancesAndPaths(t *testing.T) {
	reg, cleanup := setupTestDB(t)
	defer cleanup()

	env := &models.Env{Name: "Go", Key: "go", Category: models.CategoryRuntime}
	if err := reg.SaveEnv(env); err != nil {
		t.Fatalf("save env failed: %v", err)
	}
	inst := &models.EnvInstance{
		EnvID:       env.ID,
		Version:     "go1.25.0",
		InstallPath: `C:\Go`,
		IsDefault:   true,
		IsActive:    true,
		Source:      "system",
		DetectedAt:  time.Now(),
	}
	if err := reg.SaveInstance(inst); err != nil {
		t.Fatalf("save instance failed: %v", err)
	}
	path := &models.EnvPath{
		EnvID:     env.ID,
		Type:      models.PathInstall,
		Path:      `C:\Go`,
		SizeBytes: 123,
		IsMovable: true,
		LastSized: time.Now(),
	}
	if err := reg.SavePath(path); err != nil {
		t.Fatalf("save path failed: %v", err)
	}

	data, err := reg.ExportSnapshotData()
	if err != nil {
		t.Fatalf("export snapshot data failed: %v", err)
	}
	if err := reg.ClearInstances(env.ID); err != nil {
		t.Fatalf("clear instances failed: %v", err)
	}
	if err := reg.ClearPaths(env.ID); err != nil {
		t.Fatalf("clear paths failed: %v", err)
	}

	if err := reg.ImportSnapshotData(data); err != nil {
		t.Fatalf("import snapshot data failed: %v", err)
	}

	restoredInstances, err := reg.ListInstances(env.ID)
	if err != nil {
		t.Fatalf("list restored instances failed: %v", err)
	}
	if len(restoredInstances) != 1 || restoredInstances[0].InstallPath != inst.InstallPath {
		t.Fatalf("expected restored instance %q, got %#v", inst.InstallPath, restoredInstances)
	}
	restoredPaths, err := reg.ListPaths(env.ID)
	if err != nil {
		t.Fatalf("list restored paths failed: %v", err)
	}
	if len(restoredPaths) != 1 || restoredPaths[0].Path != path.Path {
		t.Fatalf("expected restored path %q, got %#v", path.Path, restoredPaths)
	}
}

func TestGetSettingsDefault(t *testing.T) {
	reg, cleanup := setupTestDB(t)
	defer cleanup()

	settings, err := reg.GetSettings()
	if err != nil {
		t.Fatalf("get settings failed: %v", err)
	}
	if settings.AutoScanOnStartup {
		t.Error("AutoScanOnStartup should default to false")
	}
	if !settings.ConfirmBeforeMigration {
		t.Error("ConfirmBeforeMigration should default to true")
	}
	if settings.Theme != "dark" {
		t.Errorf("Theme should default to 'dark', got %s", settings.Theme)
	}
	if len(settings.CustomScanPaths) != 0 {
		t.Errorf("CustomScanPaths should default to empty, got %v", settings.CustomScanPaths)
	}
}

func TestSaveAndGetSettings(t *testing.T) {
	reg, cleanup := setupTestDB(t)
	defer cleanup()

	settings := &models.AppSettings{
		AutoScanOnStartup:      true,
		ConfirmBeforeMigration: false,
		Theme:                  "light",
		CustomScanPaths:        []string{"/custom/path1", "/custom/path2"},
	}
	if err := reg.SaveSettings(settings); err != nil {
		t.Fatalf("save settings failed: %v", err)
	}

	fetched, err := reg.GetSettings()
	if err != nil {
		t.Fatalf("get settings failed: %v", err)
	}
	if !fetched.AutoScanOnStartup {
		t.Error("AutoScanOnStartup should be true")
	}
	if fetched.ConfirmBeforeMigration {
		t.Error("ConfirmBeforeMigration should be false")
	}
	if fetched.Theme != "light" {
		t.Errorf("Theme should be 'light', got %s", fetched.Theme)
	}
	if len(fetched.CustomScanPaths) != 2 {
		t.Errorf("expected 2 custom scan paths, got %d", len(fetched.CustomScanPaths))
	}
	if fetched.CustomScanPaths[0] != "/custom/path1" {
		t.Errorf("expected first path '/custom/path1', got %s", fetched.CustomScanPaths[0])
	}
}

func TestSaveSettingsOverwrite(t *testing.T) {
	reg, cleanup := setupTestDB(t)
	defer cleanup()

	settings := &models.AppSettings{
		AutoScanOnStartup:      true,
		ConfirmBeforeMigration: true,
		Theme:                  "dark",
		CustomScanPaths:        []string{},
	}
	if err := reg.SaveSettings(settings); err != nil {
		t.Fatalf("save settings failed: %v", err)
	}

	updated := &models.AppSettings{
		AutoScanOnStartup:      false,
		ConfirmBeforeMigration: false,
		Theme:                  "dark",
		CustomScanPaths:        []string{"/new/path"},
	}
	if err := reg.SaveSettings(updated); err != nil {
		t.Fatalf("overwrite settings failed: %v", err)
	}

	fetched, err := reg.GetSettings()
	if err != nil {
		t.Fatalf("get settings failed: %v", err)
	}
	if fetched.AutoScanOnStartup {
		t.Error("AutoScanOnStartup should be false after overwrite")
	}
	if len(fetched.CustomScanPaths) != 1 {
		t.Errorf("expected 1 custom scan path after overwrite, got %d", len(fetched.CustomScanPaths))
	}
}
