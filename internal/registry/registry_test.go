package registry

import (
	"database/sql"
	"devman/internal/models"
	"devman/internal/versionmanager"
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

func TestOpenMigratesOldEnvSchema(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "devman.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open sqlite db failed: %v", err)
	}
	if _, err := db.Exec(`
CREATE TABLE envs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    key TEXT UNIQUE NOT NULL,
    category TEXT,
    icon TEXT,
    description TEXT,
    website TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO envs (name, key, category, icon, description, website) VALUES ('Go', 'go', 'runtime', '', 'Go toolchain', 'https://go.dev');
`); err != nil {
		t.Fatalf("create old schema failed: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close old schema db failed: %v", err)
	}

	DbPathOverride = dbPath
	reg, err := Open()
	if err != nil {
		t.Fatalf("open migrated registry failed: %v", err)
	}
	defer func() {
		reg.Close()
		DbPathOverride = ""
	}()

	env, err := reg.GetEnvByKey("go")
	if err != nil {
		t.Fatalf("get migrated env failed: %v", err)
	}
	if env == nil {
		t.Fatal("expected old schema env to remain available")
	}
	if env.IsManaged {
		t.Fatal("old schema env should default to unmanaged")
	}
	managed, err := reg.SetEnvManaged("go", true)
	if err != nil {
		t.Fatalf("set migrated env managed failed: %v", err)
	}
	if !managed.IsManaged {
		t.Fatal("expected migrated env to be manageable")
	}
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

func TestSetEnvManaged(t *testing.T) {
	reg, cleanup := setupTestDB(t)
	defer cleanup()

	env := &models.Env{Name: "Go", Key: "go", Category: models.CategoryRuntime}
	if err := reg.SaveEnv(env); err != nil {
		t.Fatalf("save env failed: %v", err)
	}
	fetched, err := reg.GetEnvByKey("go")
	if err != nil {
		t.Fatalf("get env failed: %v", err)
	}
	if fetched.IsManaged {
		t.Fatal("newly saved env should default to unmanaged")
	}

	managed, err := reg.SetEnvManaged("go", true)
	if err != nil {
		t.Fatalf("set env managed failed: %v", err)
	}
	if !managed.IsManaged {
		t.Fatal("expected env to be managed")
	}

	unmanaged, err := reg.SetEnvManaged("go", false)
	if err != nil {
		t.Fatalf("set env unmanaged failed: %v", err)
	}
	if unmanaged.IsManaged {
		t.Fatal("expected env to be unmanaged")
	}
}

func TestSetEnvManagedRejectsUnknownAndEmptyKeys(t *testing.T) {
	reg, cleanup := setupTestDB(t)
	defer cleanup()

	if _, err := reg.SetEnvManaged("missing", true); err == nil {
		t.Fatal("expected unknown key to return an error")
	}
	if _, err := reg.SetEnvManaged(" ", true); err == nil {
		t.Fatal("expected empty key to return an error")
	}
}

func TestSaveEnvPreservesManagedStateOnConflict(t *testing.T) {
	reg, cleanup := setupTestDB(t)
	defer cleanup()

	env := &models.Env{Name: "Go", Key: "go", Category: models.CategoryRuntime, Description: "Original"}
	if err := reg.SaveEnv(env); err != nil {
		t.Fatalf("save env failed: %v", err)
	}
	if _, err := reg.SetEnvManaged("go", true); err != nil {
		t.Fatalf("set env managed failed: %v", err)
	}

	scanned := &models.Env{
		Name:        "Go Updated",
		Key:         "go",
		Category:    models.CategoryRuntime,
		Description: "Scanner metadata update",
		IsManaged:   false,
	}
	if err := reg.SaveEnv(scanned); err != nil {
		t.Fatalf("save scanner env failed: %v", err)
	}
	if !scanned.IsManaged {
		t.Fatal("save env should reload preserved managed state into the input env")
	}

	fetched, err := reg.GetEnvByKey("go")
	if err != nil {
		t.Fatalf("get env failed: %v", err)
	}
	if !fetched.IsManaged {
		t.Fatal("scanner save should not clear managed state")
	}
	if fetched.Name != "Go Updated" || fetched.Description != "Scanner metadata update" {
		t.Fatalf("scanner metadata should update while preserving managed state, got %#v", fetched)
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

func TestSaveAndListMetricSnapshots(t *testing.T) {
	reg, cleanup := setupTestDB(t)
	defer cleanup()

	now := time.Now()
	s1 := &models.MetricSnapshot{
		MetricKey:  "env_total_size",
		TargetKey:  "nodejs",
		ValueBytes: 1024 * 1024 * 500,
		CapturedAt: now,
	}
	if err := reg.SaveMetricSnapshot(s1); err != nil {
		t.Fatalf("save metric snapshot failed: %v", err)
	}
	if s1.ID == 0 {
		t.Error("metric snapshot ID should be assigned after insert")
	}

	s2 := &models.MetricSnapshot{
		MetricKey:  "env_total_size",
		TargetKey:  "nodejs",
		ValueBytes: 1024 * 1024 * 600,
		CapturedAt: now.Add(time.Hour),
	}
	if err := reg.SaveMetricSnapshot(s2); err != nil {
		t.Fatalf("save second metric snapshot failed: %v", err)
	}

	snapshots, err := reg.ListMetricSnapshots("env_total_size", "nodejs", 10)
	if err != nil {
		t.Fatalf("list metric snapshots failed: %v", err)
	}
	if len(snapshots) != 2 {
		t.Errorf("expected 2 snapshots, got %d", len(snapshots))
	}
	if snapshots[0].ValueBytes <= snapshots[1].ValueBytes {
		t.Error("snapshots should be ordered by captured_at DESC")
	}

	allSnapshots, err := reg.ListMetricSnapshots("env_total_size", "", 10)
	if err != nil {
		t.Fatalf("list all metric snapshots failed: %v", err)
	}
	if len(allSnapshots) != 2 {
		t.Errorf("expected 2 snapshots without target filter, got %d", len(allSnapshots))
	}
}

func TestListMetricSnapshotsDifferentKeys(t *testing.T) {
	reg, cleanup := setupTestDB(t)
	defer cleanup()

	now := time.Now()
	sizeSnap := &models.MetricSnapshot{
		MetricKey:  "env_total_size",
		TargetKey:  "go",
		ValueBytes: 300 * 1024 * 1024,
		CapturedAt: now,
	}
	cleanSnap := &models.MetricSnapshot{
		MetricKey:  "clean_freed_bytes",
		TargetKey:  "all",
		ValueBytes: 50 * 1024 * 1024,
		CapturedAt: now,
	}
	if err := reg.SaveMetricSnapshot(sizeSnap); err != nil {
		t.Fatalf("save size snapshot failed: %v", err)
	}
	if err := reg.SaveMetricSnapshot(cleanSnap); err != nil {
		t.Fatalf("save clean snapshot failed: %v", err)
	}

	sizeList, err := reg.ListMetricSnapshots("env_total_size", "", 10)
	if err != nil {
		t.Fatalf("list size snapshots failed: %v", err)
	}
	if len(sizeList) != 1 {
		t.Errorf("expected 1 size snapshot, got %d", len(sizeList))
	}

	cleanList, err := reg.ListMetricSnapshots("clean_freed_bytes", "", 10)
	if err != nil {
		t.Fatalf("list clean snapshots failed: %v", err)
	}
	if len(cleanList) != 1 {
		t.Errorf("expected 1 clean snapshot, got %d", len(cleanList))
	}
}

func TestAggregateMetricSnapshot(t *testing.T) {
	reg, cleanup := setupTestDB(t)
	defer cleanup()

	now := time.Now()
	perEnvSnapshots := []struct {
		targetKey  string
		valueBytes int64
	}{
		{"nodejs", 500 * 1024 * 1024},
		{"python", 300 * 1024 * 1024},
		{"go", 200 * 1024 * 1024},
	}
	var totalSize int64
	for i, s := range perEnvSnapshots {
		totalSize += s.valueBytes
		if err := reg.SaveMetricSnapshot(&models.MetricSnapshot{
			MetricKey:  "env_total_size",
			TargetKey:  s.targetKey,
			ValueBytes: s.valueBytes,
			CapturedAt: now,
		}); err != nil {
			t.Fatalf("save per-env snapshot %d failed: %v", i, err)
		}
	}
	if err := reg.SaveMetricSnapshot(&models.MetricSnapshot{
		MetricKey:  "env_total_size",
		TargetKey:  "all",
		ValueBytes: totalSize,
		CapturedAt: now,
	}); err != nil {
		t.Fatalf("save aggregate snapshot failed: %v", err)
	}

	aggSnapshots, err := reg.ListMetricSnapshots("env_total_size", "all", 10)
	if err != nil {
		t.Fatalf("list aggregate snapshots failed: %v", err)
	}
	if len(aggSnapshots) != 1 {
		t.Fatalf("expected 1 aggregate snapshot, got %d", len(aggSnapshots))
	}
	if aggSnapshots[0].ValueBytes != totalSize {
		t.Errorf("expected aggregate value %d, got %d", totalSize, aggSnapshots[0].ValueBytes)
	}
	if aggSnapshots[0].TargetKey != "all" {
		t.Errorf("expected target_key 'all', got %s", aggSnapshots[0].TargetKey)
	}
}

func TestPruneMetricSnapshots(t *testing.T) {
	reg, cleanup := setupTestDB(t)
	defer cleanup()

	now := time.Now()
	for i := 0; i < 5; i++ {
		s := &models.MetricSnapshot{
			MetricKey:  "env_total_size",
			TargetKey:  "nodejs",
			ValueBytes: int64((i + 1) * 100 * 1024 * 1024),
			CapturedAt: now.Add(time.Duration(i) * time.Hour),
		}
		if err := reg.SaveMetricSnapshot(s); err != nil {
			t.Fatalf("save snapshot %d failed: %v", i, err)
		}
	}

	if err := reg.PruneMetricSnapshots("env_total_size", 3); err != nil {
		t.Fatalf("prune metric snapshots failed: %v", err)
	}

	snapshots, err := reg.ListMetricSnapshots("env_total_size", "nodejs", 10)
	if err != nil {
		t.Fatalf("list after prune failed: %v", err)
	}
	if len(snapshots) != 3 {
		t.Errorf("expected 3 snapshots after prune, got %d", len(snapshots))
	}
}

func TestVersionManagementTablesPersistRecords(t *testing.T) {
	reg, cleanup := setupTestDB(t)
	defer cleanup()

	version := &versionmanager.ManagedVersion{
		ToolKey:      "go",
		Version:      "1.26.0",
		InstallPath:  `D:\production\go1.26`,
		BinPath:      `D:\production\go1.26\bin\go.exe`,
		Source:       versionmanager.SourceDevMan,
		IsDefault:    true,
		IsActive:     true,
		CanDelete:    true,
		DeletePolicy: versionmanager.DeletePolicyDirect,
		DetectedAt:   time.Now(),
	}
	if err := reg.SaveToolVersion(version); err != nil {
		t.Fatalf("SaveToolVersion failed: %v", err)
	}
	if version.ID == 0 {
		t.Fatal("expected SaveToolVersion to assign ID")
	}

	versions, err := reg.ListToolVersions("go")
	if err != nil {
		t.Fatalf("ListToolVersions failed: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("expected 1 version, got %#v", versions)
	}
	if versions[0].ToolKey != "go" || versions[0].Version != "1.26.0" {
		t.Fatalf("unexpected version: %#v", versions[0])
	}

	strategy := versionmanager.InstallStrategy{
		ToolKey:   "go",
		RootDir:   `D:\production`,
		Reason:    "confirmed by user",
		UpdatedAt: time.Now(),
	}
	if err := reg.SaveInstallStrategy(strategy); err != nil {
		t.Fatalf("SaveInstallStrategy failed: %v", err)
	}
	gotStrategy, err := reg.GetInstallStrategy("go")
	if err != nil {
		t.Fatalf("GetInstallStrategy failed: %v", err)
	}
	if gotStrategy == nil || gotStrategy.RootDir != `D:\production` {
		t.Fatalf("unexpected strategy: %#v", gotStrategy)
	}
}

func TestSaveToolVersionUpsertReloadsExistingIDAndUpdatesFields(t *testing.T) {
	reg, cleanup := setupTestDB(t)
	defer cleanup()

	first := &versionmanager.ManagedVersion{
		ToolKey:      "go",
		Version:      "1.25.0",
		InstallPath:  `D:\production\go`,
		BinPath:      `D:\production\go\bin\go.exe`,
		Source:       versionmanager.SourceDevMan,
		IsDefault:    false,
		IsActive:     true,
		CanDelete:    false,
		DeletePolicy: versionmanager.DeletePolicyBlocked,
		DetectedAt:   time.Now(),
	}
	if err := reg.SaveToolVersion(first); err != nil {
		t.Fatalf("SaveToolVersion first failed: %v", err)
	}
	if first.ID == 0 {
		t.Fatal("expected first save to assign ID")
	}

	other := &versionmanager.ManagedVersion{
		ToolKey:      "node",
		Version:      "24.0.0",
		InstallPath:  `D:\production\node`,
		Source:       versionmanager.SourceDevMan,
		DeletePolicy: versionmanager.DeletePolicyDirect,
		DetectedAt:   time.Now(),
	}
	if err := reg.SaveToolVersion(other); err != nil {
		t.Fatalf("SaveToolVersion other failed: %v", err)
	}

	second := &versionmanager.ManagedVersion{
		ToolKey:      "go",
		Version:      "1.26.0",
		InstallPath:  `D:\production\go`,
		BinPath:      `D:\production\go\bin\go.exe`,
		Source:       versionmanager.SourceDevMan,
		IsDefault:    true,
		IsActive:     true,
		CanDelete:    true,
		DeletePolicy: versionmanager.DeletePolicyDirect,
		DetectedAt:   time.Now(),
	}
	if err := reg.SaveToolVersion(second); err != nil {
		t.Fatalf("SaveToolVersion second failed: %v", err)
	}
	if second.ID != first.ID {
		t.Fatalf("expected upsert to reload existing id %d, got %d", first.ID, second.ID)
	}
	var rowCount int
	if err := reg.db.QueryRow(`SELECT COUNT(*) FROM tool_versions WHERE tool_key = ? AND install_path = ?`, "go", `D:\production\go`).Scan(&rowCount); err != nil {
		t.Fatalf("count upserted rows failed: %v", err)
	}
	if rowCount != 1 {
		t.Fatalf("expected exactly one row for upserted path, got %d", rowCount)
	}

	versions, err := reg.ListToolVersions("go")
	if err != nil {
		t.Fatalf("ListToolVersions failed: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("expected exactly one go version after upsert, got %#v", versions)
	}
	got := versions[0]
	if got.ID != first.ID || got.Version != "1.26.0" || !got.IsDefault || !got.CanDelete || got.DeletePolicy != versionmanager.DeletePolicyDirect {
		t.Fatalf("upsert fields were not persisted: %#v", got)
	}
}

func TestListToolVersionsToleratesNullableOptionalFields(t *testing.T) {
	reg, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := reg.db.Exec(`
		INSERT INTO tool_versions (
			tool_key, version, install_path, bin_path, source,
			is_default, is_active, can_delete, delete_policy, detected_at
		) VALUES (?, ?, ?, NULL, ?, ?, ?, ?, NULL, ?)
	`, "go", "1.26.0", `D:\production\go-nullable`, string(versionmanager.SourceExternal), 0, 1, 0, time.Now())
	if err != nil {
		t.Fatalf("insert nullable tool version failed: %v", err)
	}

	versions, err := reg.ListToolVersions("go")
	if err != nil {
		t.Fatalf("ListToolVersions should tolerate nullable fields: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("expected 1 version, got %#v", versions)
	}
	if versions[0].BinPath != "" || versions[0].DeletePolicy != "" {
		t.Fatalf("expected empty optional fields, got %#v", versions[0])
	}
}

func TestInstallStrategyNullableReasonAndUpsertReplacement(t *testing.T) {
	reg, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := reg.db.Exec(`
		INSERT INTO version_install_strategies (tool_key, root_dir, reason, updated_at)
		VALUES (?, ?, NULL, ?)
	`, "go", `D:\production`, time.Now())
	if err != nil {
		t.Fatalf("insert nullable install strategy failed: %v", err)
	}

	got, err := reg.GetInstallStrategy("go")
	if err != nil {
		t.Fatalf("GetInstallStrategy should tolerate nullable reason: %v", err)
	}
	if got == nil || got.Reason != "" {
		t.Fatalf("expected empty reason, got %#v", got)
	}

	replacement := versionmanager.InstallStrategy{
		ToolKey:   "go",
		RootDir:   `D:\tools`,
		Reason:    "changed by user",
		UpdatedAt: time.Now(),
	}
	if err := reg.SaveInstallStrategy(replacement); err != nil {
		t.Fatalf("SaveInstallStrategy replacement failed: %v", err)
	}
	got, err = reg.GetInstallStrategy("go")
	if err != nil {
		t.Fatalf("GetInstallStrategy replacement failed: %v", err)
	}
	if got == nil || got.RootDir != `D:\tools` || got.Reason != "changed by user" {
		t.Fatalf("expected replacement strategy, got %#v", got)
	}
}
