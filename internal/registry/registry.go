package registry

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"devman/internal/models"
	"devman/internal/versionmanager"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

type Registry struct {
	db *sql.DB
}

var DbPathOverride string

func dbPath() string {
	if DbPathOverride != "" {
		return DbPathOverride
	}
	exe, err := os.Executable()
	if err != nil {
		return "devman.db"
	}
	return filepath.Join(filepath.Dir(exe), "devman.db")
}

func logPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "devman.log"
	}
	return filepath.Join(filepath.Dir(exe), "devman.log")
}

func Open() (*Registry, error) {
	path := dbPath()
	logrus.WithField("db_path", path).Info("opening registry database")
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		logrus.WithError(err).WithField("db_path", path).Error("failed to open registry database")
		return nil, err
	}
	if err := db.Ping(); err != nil {
		logrus.WithError(err).WithField("db_path", path).Error("failed to ping registry database")
		return nil, err
	}
	r := &Registry{db: db}
	if err := r.migrate(); err != nil {
		logrus.WithError(err).WithField("db_path", path).Error("failed to migrate registry database")
		return nil, err
	}
	logrus.WithField("db_path", path).Info("registry database opened")
	return r, nil
}

func (r *Registry) Close() error {
	logrus.Info("closing registry database")
	return r.db.Close()
}

func (r *Registry) migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS envs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    key TEXT UNIQUE NOT NULL,
    category TEXT,
    icon TEXT,
    description TEXT,
    website TEXT,
    is_managed INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS env_instances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    env_id INTEGER REFERENCES envs(id),
    version TEXT NOT NULL,
    install_path TEXT NOT NULL,
    is_default INTEGER DEFAULT 0,
    is_active INTEGER DEFAULT 1,
    source TEXT DEFAULT 'system',
    detected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS env_paths (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    env_id INTEGER REFERENCES envs(id),
    instance_id INTEGER REFERENCES env_instances(id),
    type TEXT NOT NULL,
    path TEXT NOT NULL,
    size_bytes INTEGER DEFAULT 0,
    is_movable INTEGER DEFAULT 1,
    last_sized TIMESTAMP,
    UNIQUE(env_id, type, path)
);

CREATE TABLE IF NOT EXISTS snapshots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    data_json TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    action TEXT NOT NULL,
    target_env TEXT,
    details_json TEXT,
    success INTEGER DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_envs_key ON envs(key);
CREATE INDEX IF NOT EXISTS idx_instances_env ON env_instances(env_id);
CREATE INDEX IF NOT EXISTS idx_paths_env ON env_paths(env_id);

CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value_json TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS metric_snapshots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    metric_key TEXT NOT NULL,
    target_key TEXT NOT NULL,
    value_bytes INTEGER NOT NULL,
    captured_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_metrics_key ON metric_snapshots(metric_key);
CREATE INDEX IF NOT EXISTS idx_metrics_target ON metric_snapshots(metric_key, target_key);
CREATE INDEX IF NOT EXISTS idx_metrics_captured ON metric_snapshots(captured_at);

CREATE TABLE IF NOT EXISTS tool_versions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tool_key TEXT NOT NULL,
    version TEXT NOT NULL,
    install_path TEXT NOT NULL,
    bin_path TEXT,
    source TEXT NOT NULL,
    is_default INTEGER DEFAULT 0,
    is_active INTEGER DEFAULT 0,
    can_delete INTEGER DEFAULT 0,
    delete_policy TEXT,
    detected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tool_key, install_path)
);

CREATE TABLE IF NOT EXISTS version_install_strategies (
    tool_key TEXT PRIMARY KEY,
    root_dir TEXT NOT NULL,
    reason TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS version_operations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tool_key TEXT NOT NULL,
    version TEXT,
    operation TEXT NOT NULL,
    success INTEGER DEFAULT 0,
    message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tool_versions_tool ON tool_versions(tool_key);
CREATE INDEX IF NOT EXISTS idx_version_operations_tool ON version_operations(tool_key);
`
	_, err := r.db.Exec(schema)
	if err != nil {
		logrus.WithError(err).Error("registry schema migration failed")
		return err
	}
	if err := r.ensureEnvColumns(); err != nil {
		logrus.WithError(err).Error("registry env column migration failed")
		return err
	}
	logrus.Info("registry schema migration complete")
	return err
}

func (r *Registry) ensureEnvColumns() error {
	hasManaged, err := r.hasColumn("envs", "is_managed")
	if err != nil {
		return err
	}
	if hasManaged {
		return nil
	}
	_, err = r.db.Exec(`ALTER TABLE envs ADD COLUMN is_managed INTEGER DEFAULT 0`)
	return err
}

func (r *Registry) hasColumn(table string, column string) (bool, error) {
	rows, err := r.db.Query(fmt.Sprintf(`PRAGMA table_info(%s)`, table))
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull, pk int
		var defaultValue interface{}
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}

func (r *Registry) SaveEnv(env *models.Env) error {
	now := time.Now()
	if env.CreatedAt.IsZero() {
		env.CreatedAt = now
	}
	env.UpdatedAt = now

	_, err := r.db.Exec(
		`INSERT INTO envs (name, key, category, icon, description, website, is_managed, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(key) DO UPDATE SET
		   name=excluded.name, category=excluded.category, icon=excluded.icon,
		   description=excluded.description, website=excluded.website,
		   is_managed=envs.is_managed, updated_at=excluded.updated_at`,
		env.Name, env.Key, env.Category, env.Icon, env.Description, env.Website, env.IsManaged,
		env.CreatedAt, env.UpdatedAt,
	)
	if err != nil {
		logrus.WithError(err).WithField("env_key", env.Key).Error("failed to save environment")
		return err
	}
	saved, err := r.GetEnvByKey(env.Key)
	if err != nil {
		logrus.WithError(err).WithField("env_key", env.Key).Error("failed to reload saved environment")
		return err
	}
	if saved == nil {
		return fmt.Errorf("env not found after save: %s", env.Key)
	}
	*env = *saved
	return nil
}

func (r *Registry) SetEnvManaged(key string, managed bool) (*models.Env, error) {
	if strings.TrimSpace(key) == "" {
		return nil, fmt.Errorf("env key is required")
	}
	now := time.Now()
	managedValue := 0
	if managed {
		managedValue = 1
	}
	res, err := r.db.Exec(`UPDATE envs SET is_managed = ?, updated_at = ? WHERE key = ?`, managedValue, now, key)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"env_key": key, "managed": managed}).Error("failed to update environment management state")
		return nil, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		logrus.WithError(err).WithField("env_key", key).Error("failed to inspect environment management update")
		return nil, err
	}
	if affected == 0 {
		return nil, fmt.Errorf("env not found: %s", key)
	}
	env, err := r.GetEnvByKey(key)
	if err != nil {
		logrus.WithError(err).WithField("env_key", key).Error("failed to reload managed environment")
		return nil, err
	}
	if env == nil {
		return nil, fmt.Errorf("env not found: %s", key)
	}
	return env, nil
}

func (r *Registry) GetEnvByKey(key string) (*models.Env, error) {
	row := r.db.QueryRow(`SELECT id, name, key, category, icon, description, website, is_managed, created_at, updated_at FROM envs WHERE key = ?`, key)
	e := &models.Env{}
	var im int
	err := row.Scan(&e.ID, &e.Name, &e.Key, &e.Category, &e.Icon, &e.Description, &e.Website, &im, &e.CreatedAt, &e.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	e.IsManaged = im != 0
	return e, nil
}

func (r *Registry) ListEnvs() ([]models.Env, error) {
	rows, err := r.db.Query(`SELECT id, name, key, category, icon, description, website, is_managed, created_at, updated_at FROM envs ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Env
	for rows.Next() {
		e := models.Env{}
		var im int
		if err := rows.Scan(&e.ID, &e.Name, &e.Key, &e.Category, &e.Icon, &e.Description, &e.Website, &im, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		e.IsManaged = im != 0
		list = append(list, e)
	}
	return list, rows.Err()
}

func (r *Registry) SaveInstance(inst *models.EnvInstance) error {
	res, err := r.db.Exec(
		`INSERT INTO env_instances (env_id, version, install_path, is_default, is_active, source, detected_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		inst.EnvID, inst.Version, inst.InstallPath, inst.IsDefault, inst.IsActive, inst.Source, inst.DetectedAt,
	)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"env_id": inst.EnvID, "install_path": inst.InstallPath}).Error("failed to save environment instance")
		return err
	}
	id, _ := res.LastInsertId()
	inst.ID = id
	return nil
}

func (r *Registry) ClearInstances(envID int64) error {
	_, err := r.db.Exec(`DELETE FROM env_instances WHERE env_id = ?`, envID)
	return err
}

// DeleteEnv removes the env row identified by key along with its dependent
// env_instances and env_paths in a single transaction. It is idempotent: when
// no row matches the key (already deleted, never persisted, or empty key), it
// returns nil without modifying the database. History, metric snapshots, and
// tool_versions are intentionally untouched.
func (r *Registry) DeleteEnv(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil
	}
	tx, err := r.db.Begin()
	if err != nil {
		logrus.WithError(err).WithField("env_key", key).Error("failed to begin env deletion transaction")
		return err
	}
	defer tx.Rollback()

	var envID int64
	if err := tx.QueryRow(`SELECT id FROM envs WHERE key = ?`, key).Scan(&envID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		logrus.WithError(err).WithField("env_key", key).Error("failed to look up env for deletion")
		return err
	}

	if _, err := tx.Exec(`DELETE FROM env_instances WHERE env_id = ?`, envID); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"env_key": key, "env_id": envID}).Error("failed to delete env instances")
		return err
	}
	if _, err := tx.Exec(`DELETE FROM env_paths WHERE env_id = ?`, envID); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"env_key": key, "env_id": envID}).Error("failed to delete env paths")
		return err
	}
	if _, err := tx.Exec(`DELETE FROM envs WHERE id = ?`, envID); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"env_key": key, "env_id": envID}).Error("failed to delete env row")
		return err
	}

	if err := tx.Commit(); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"env_key": key, "env_id": envID}).Error("failed to commit env deletion")
		return err
	}
	return nil
}

func (r *Registry) ListInstances(envID int64) ([]models.EnvInstance, error) {
	rows, err := r.db.Query(`SELECT id, env_id, version, install_path, is_default, is_active, source, detected_at FROM env_instances WHERE env_id = ?`, envID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.EnvInstance
	for rows.Next() {
		var i models.EnvInstance
		var def, act int
		if err := rows.Scan(&i.ID, &i.EnvID, &i.Version, &i.InstallPath, &def, &act, &i.Source, &i.DetectedAt); err != nil {
			return nil, err
		}
		i.IsDefault = def != 0
		i.IsActive = act != 0
		list = append(list, i)
	}
	return list, rows.Err()
}

func (r *Registry) SavePath(p *models.EnvPath) error {
	res, err := r.db.Exec(
		`INSERT INTO env_paths (env_id, instance_id, type, path, size_bytes, is_movable, last_sized)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(env_id, type, path) DO UPDATE SET
		   size_bytes=excluded.size_bytes, is_movable=excluded.is_movable, last_sized=excluded.last_sized`,
		p.EnvID, p.InstanceID, p.Type, p.Path, p.SizeBytes, p.IsMovable, p.LastSized,
	)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"env_id": p.EnvID, "type": p.Type, "path": p.Path}).Error("failed to save environment path")
		return err
	}
	id, _ := res.LastInsertId()
	p.ID = id
	return nil
}

func (r *Registry) ClearPaths(envID int64) error {
	_, err := r.db.Exec(`DELETE FROM env_paths WHERE env_id = ?`, envID)
	return err
}

func (r *Registry) ListPaths(envID int64) ([]models.EnvPath, error) {
	rows, err := r.db.Query(`SELECT id, env_id, instance_id, type, path, size_bytes, is_movable, last_sized FROM env_paths WHERE env_id = ?`, envID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.EnvPath
	for rows.Next() {
		var p models.EnvPath
		var im int
		if err := rows.Scan(&p.ID, &p.EnvID, &p.InstanceID, &p.Type, &p.Path, &p.SizeBytes, &im, &p.LastSized); err != nil {
			return nil, err
		}
		p.IsMovable = im != 0
		list = append(list, p)
	}
	return list, rows.Err()
}

func (r *Registry) SaveSnapshot(s *models.Snapshot) error {
	res, err := r.db.Exec(`INSERT INTO snapshots (name, data_json, created_at) VALUES (?, ?, ?)`, s.Name, s.DataJSON, s.CreatedAt)
	if err != nil {
		logrus.WithError(err).WithField("snapshot_name", s.Name).Error("failed to save snapshot")
		return err
	}
	id, _ := res.LastInsertId()
	s.ID = id
	return nil
}

func (r *Registry) SaveHistory(h *models.HistoryEntry) error {
	res, err := r.db.Exec(
		`INSERT INTO history (action, target_env, details_json, success, error_message, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		h.Action, h.TargetEnv, h.DetailsJSON, h.Success, h.ErrorMessage, h.CreatedAt,
	)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"action": h.Action, "target_env": h.TargetEnv}).Error("failed to save history")
		return err
	}
	id, _ := res.LastInsertId()
	h.ID = id
	return nil
}

func (r *Registry) GetHistory(limit int) ([]models.HistoryEntry, error) {
	rows, err := r.db.Query(`SELECT id, action, target_env, details_json, success, error_message, created_at FROM history ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.HistoryEntry
	for rows.Next() {
		var h models.HistoryEntry
		var s int
		if err := rows.Scan(&h.ID, &h.Action, &h.TargetEnv, &h.DetailsJSON, &s, &h.ErrorMessage, &h.CreatedAt); err != nil {
			return nil, err
		}
		h.Success = s != 0
		list = append(list, h)
	}
	return list, rows.Err()
}

func (r *Registry) GetTotalSizeByEnv() (map[int64]int64, error) {
	rows, err := r.db.Query(`SELECT env_id, COALESCE(SUM(size_bytes), 0) FROM env_paths GROUP BY env_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[int64]int64)
	for rows.Next() {
		var eid int64
		var size int64
		if err := rows.Scan(&eid, &size); err != nil {
			return nil, err
		}
		m[eid] = size
	}
	return m, rows.Err()
}

func (r *Registry) GetSnapshot(id int64) (*models.Snapshot, error) {
	row := r.db.QueryRow(`SELECT id, name, data_json, created_at FROM snapshots WHERE id = ?`, id)
	s := &models.Snapshot{}
	err := row.Scan(&s.ID, &s.Name, &s.DataJSON, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Registry) DeletePathsNotIn(envID int64, paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	// Build placeholders
	placeholders := make([]string, len(paths))
	args := make([]interface{}, len(paths)+1)
	args[0] = envID
	for i, p := range paths {
		placeholders[i] = "?"
		args[i+1] = p
	}
	q := fmt.Sprintf(`DELETE FROM env_paths WHERE env_id = ? AND path NOT IN (%s)`, joinPlaceholders(placeholders))
	_, err := r.db.Exec(q, args...)
	return err
}

func joinPlaceholders(p []string) string {
	if len(p) == 0 {
		return ""
	}
	result := p[0]
	for i := 1; i < len(p); i++ {
		result += ", " + p[i]
	}
	return result
}

// Export/Import for snapshots
func (r *Registry) GetSettings() (*models.AppSettings, error) {
	row := r.db.QueryRow(`SELECT value_json FROM settings WHERE key = 'app'`)
	var jsonStr string
	err := row.Scan(&jsonStr)
	if err == sql.ErrNoRows {
		// Return defaults
		return &models.AppSettings{
			AutoScanOnStartup:      false,
			ConfirmBeforeMigration: true,
			Theme:                  "dark",
			CustomScanPaths:        []string{},
		}, nil
	}
	if err != nil {
		return nil, err
	}
	var s models.AppSettings
	if err := json.Unmarshal([]byte(jsonStr), &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Registry) SaveSettings(s *models.AppSettings) error {
	data, err := json.Marshal(s)
	if err != nil {
		logrus.WithError(err).Error("failed to marshal settings")
		return err
	}
	_, err = r.db.Exec(
		`INSERT INTO settings (key, value_json, updated_at) VALUES ('app', ?, ?)
		 ON CONFLICT(key) DO UPDATE SET value_json=excluded.value_json, updated_at=excluded.updated_at`,
		string(data), time.Now(),
	)
	if err != nil {
		logrus.WithError(err).Error("failed to save settings")
	}
	return err
}

// Export/Import for snapshots

func (r *Registry) SaveMetricSnapshot(s *models.MetricSnapshot) error {
	res, err := r.db.Exec(
		`INSERT INTO metric_snapshots (metric_key, target_key, value_bytes, captured_at) VALUES (?, ?, ?, ?)`,
		s.MetricKey, s.TargetKey, s.ValueBytes, s.CapturedAt,
	)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"metric_key": s.MetricKey, "target_key": s.TargetKey}).Error("failed to save metric snapshot")
		return err
	}
	id, _ := res.LastInsertId()
	s.ID = id
	return nil
}

func (r *Registry) ListMetricSnapshots(metricKey, targetKey string, limit int) ([]models.MetricSnapshot, error) {
	var rows *sql.Rows
	var err error
	if targetKey != "" {
		rows, err = r.db.Query(
			`SELECT id, metric_key, target_key, value_bytes, captured_at FROM metric_snapshots WHERE metric_key = ? AND target_key = ? ORDER BY captured_at DESC LIMIT ?`,
			metricKey, targetKey, limit,
		)
	} else {
		rows, err = r.db.Query(
			`SELECT id, metric_key, target_key, value_bytes, captured_at FROM metric_snapshots WHERE metric_key = ? ORDER BY captured_at DESC LIMIT ?`,
			metricKey, limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.MetricSnapshot
	for rows.Next() {
		var s models.MetricSnapshot
		if err := rows.Scan(&s.ID, &s.MetricKey, &s.TargetKey, &s.ValueBytes, &s.CapturedAt); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func (r *Registry) PruneMetricSnapshots(metricKey string, keepCount int) error {
	_, err := r.db.Exec(
		`DELETE FROM metric_snapshots WHERE metric_key = ? AND id NOT IN (SELECT id FROM metric_snapshots WHERE metric_key = ? ORDER BY captured_at DESC LIMIT ?)`,
		metricKey, metricKey, keepCount,
	)
	if err != nil {
		logrus.WithError(err).WithField("metric_key", metricKey).Error("failed to prune metric snapshots")
	}
	return err
}

func (r *Registry) SaveToolVersion(v *versionmanager.ManagedVersion) error {
	if v.DetectedAt.IsZero() {
		v.DetectedAt = time.Now()
	}
	_, err := r.db.Exec(`
		INSERT INTO tool_versions (
			tool_key, version, install_path, bin_path, source, is_default,
			is_active, can_delete, delete_policy, detected_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(tool_key, install_path) DO UPDATE SET
			version=excluded.version,
			bin_path=excluded.bin_path,
			source=excluded.source,
			is_default=excluded.is_default,
			is_active=excluded.is_active,
			can_delete=excluded.can_delete,
			delete_policy=excluded.delete_policy,
			detected_at=excluded.detected_at
	`, v.ToolKey, v.Version, v.InstallPath, v.BinPath, string(v.Source), boolToInt(v.IsDefault),
		boolToInt(v.IsActive), boolToInt(v.CanDelete), string(v.DeletePolicy), v.DetectedAt)
	if err != nil {
		return err
	}
	return r.db.QueryRow(
		`SELECT id FROM tool_versions WHERE tool_key = ? AND install_path = ?`,
		v.ToolKey, v.InstallPath,
	).Scan(&v.ID)
}

func (r *Registry) SyncScannedToolVersions(toolKey string, scanned []versionmanager.ManagedVersion) error {
	existing, err := r.ListToolVersions(toolKey)
	if err != nil {
		return err
	}

	existingByPath := make(map[string]versionmanager.ManagedVersion, len(existing))
	for _, version := range existing {
		existingByPath[version.InstallPath] = version
	}

	hasPinnedState := false
	for _, version := range existing {
		if version.Source != versionmanager.SourceExternal || version.IsDefault || version.IsActive {
			hasPinnedState = true
			break
		}
	}

	detectedPaths := make(map[string]struct{}, len(scanned))
	for _, detected := range scanned {
		installPath := strings.TrimSpace(detected.InstallPath)
		if installPath == "" {
			continue
		}
		detected.InstallPath = installPath
		if detected.DetectedAt.IsZero() {
			detected.DetectedAt = time.Now()
		}
		detectedPaths[installPath] = struct{}{}

		if current, ok := existingByPath[installPath]; ok {
			current.Version = detected.Version
			if strings.TrimSpace(detected.BinPath) != "" {
				current.BinPath = detected.BinPath
			}
			current.DetectedAt = detected.DetectedAt
			if err := r.SaveToolVersion(&current); err != nil {
				return err
			}
			continue
		}

		if hasPinnedState {
			detected.IsDefault = false
			detected.IsActive = false
		}
		if err := r.SaveToolVersion(&detected); err != nil {
			return err
		}
	}

	for _, current := range existing {
		if current.Source != versionmanager.SourceExternal {
			continue
		}
		if current.IsDefault || current.IsActive {
			continue
		}
		if _, ok := detectedPaths[current.InstallPath]; ok {
			continue
		}
		if err := r.DeleteToolVersion(current.ID); err != nil {
			return err
		}
	}

	return nil
}

func (r *Registry) ListToolVersions(toolKey string) ([]versionmanager.ManagedVersion, error) {
	rows, err := r.db.Query(`
		SELECT id, tool_key, version, install_path, COALESCE(bin_path, ''), source,
		       is_default, is_active, can_delete, COALESCE(delete_policy, ''), detected_at
		FROM tool_versions
		WHERE tool_key = ?
		ORDER BY is_default DESC, version DESC
	`, toolKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []versionmanager.ManagedVersion
	for rows.Next() {
		var v versionmanager.ManagedVersion
		var isDefault, isActive, canDelete int
		var source, deletePolicy string
		if err := rows.Scan(&v.ID, &v.ToolKey, &v.Version, &v.InstallPath, &v.BinPath, &source, &isDefault, &isActive, &canDelete, &deletePolicy, &v.DetectedAt); err != nil {
			return nil, err
		}
		v.Source = versionmanager.VersionSource(source)
		v.IsDefault = isDefault != 0
		v.IsActive = isActive != 0
		v.CanDelete = canDelete != 0
		v.DeletePolicy = versionmanager.DeletePolicy(deletePolicy)
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

func (r *Registry) DeleteToolVersion(id int64) error {
	_, err := r.db.Exec(`DELETE FROM tool_versions WHERE id = ?`, id)
	return err
}

func (r *Registry) SaveInstallStrategy(strategy versionmanager.InstallStrategy) error {
	if strategy.UpdatedAt.IsZero() {
		strategy.UpdatedAt = time.Now()
	}
	_, err := r.db.Exec(`
		INSERT INTO version_install_strategies (tool_key, root_dir, reason, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(tool_key) DO UPDATE SET
			root_dir=excluded.root_dir,
			reason=excluded.reason,
			updated_at=excluded.updated_at
	`, strategy.ToolKey, strategy.RootDir, strategy.Reason, strategy.UpdatedAt)
	return err
}

func (r *Registry) GetInstallStrategy(toolKey string) (*versionmanager.InstallStrategy, error) {
	row := r.db.QueryRow(`
		SELECT tool_key, root_dir, COALESCE(reason, ''), updated_at
		FROM version_install_strategies
		WHERE tool_key = ?
	`, toolKey)
	var strategy versionmanager.InstallStrategy
	if err := row.Scan(&strategy.ToolKey, &strategy.RootDir, &strategy.Reason, &strategy.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &strategy, nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func (r *Registry) ExportSnapshotData() (map[string]interface{}, error) {
	envs, err := r.ListEnvs()
	if err != nil {
		return nil, err
	}
	instances, err := r.db.Query(`SELECT id, env_id, version, install_path, is_default, is_active, source, detected_at FROM env_instances`)
	if err != nil {
		return nil, err
	}
	defer instances.Close()

	var instList []models.EnvInstance
	for instances.Next() {
		var i models.EnvInstance
		var def, act int
		if err := instances.Scan(&i.ID, &i.EnvID, &i.Version, &i.InstallPath, &def, &act, &i.Source, &i.DetectedAt); err != nil {
			return nil, err
		}
		i.IsDefault = def != 0
		i.IsActive = act != 0
		instList = append(instList, i)
	}

	paths, err := r.db.Query(`SELECT id, env_id, instance_id, type, path, size_bytes, is_movable, last_sized FROM env_paths`)
	if err != nil {
		return nil, err
	}
	defer paths.Close()

	var pathList []models.EnvPath
	for paths.Next() {
		var p models.EnvPath
		var im int
		if err := paths.Scan(&p.ID, &p.EnvID, &p.InstanceID, &p.Type, &p.Path, &p.SizeBytes, &im, &p.LastSized); err != nil {
			return nil, err
		}
		p.IsMovable = im != 0
		pathList = append(pathList, p)
	}

	return map[string]interface{}{
		"envs":       envs,
		"instances":  instList,
		"paths":      pathList,
		"exportedAt": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (r *Registry) ImportSnapshotData(data map[string]interface{}) error {
	envs, err := snapshotSlice[models.Env](data, "envs")
	if err != nil {
		return err
	}
	instances, err := snapshotSlice[models.EnvInstance](data, "instances")
	if err != nil {
		return err
	}
	paths, err := snapshotSlice[models.EnvPath](data, "paths")
	if err != nil {
		return err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM env_paths`); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM env_instances`); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM envs`); err != nil {
		return err
	}

	for _, e := range envs {
		if e.Key == "" {
			continue
		}
		if _, err := tx.Exec(
			`INSERT INTO envs (id, name, key, category, icon, description, website, is_managed, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			e.ID, e.Name, e.Key, e.Category, e.Icon, e.Description, e.Website, e.IsManaged, e.CreatedAt, e.UpdatedAt,
		); err != nil {
			return err
		}
	}

	for _, i := range instances {
		if i.EnvID == 0 || i.InstallPath == "" {
			continue
		}
		if _, err := tx.Exec(
			`INSERT INTO env_instances (id, env_id, version, install_path, is_default, is_active, source, detected_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			i.ID, i.EnvID, i.Version, i.InstallPath, i.IsDefault, i.IsActive, i.Source, i.DetectedAt,
		); err != nil {
			return err
		}
	}

	for _, p := range paths {
		if p.EnvID == 0 || p.Path == "" {
			continue
		}
		if _, err := tx.Exec(
			`INSERT INTO env_paths (id, env_id, instance_id, type, path, size_bytes, is_movable, last_sized)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			p.ID, p.EnvID, p.InstanceID, p.Type, p.Path, p.SizeBytes, p.IsMovable, p.LastSized,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func snapshotSlice[T any](data map[string]interface{}, key string) ([]T, error) {
	raw, ok := data[key]
	if !ok {
		return nil, nil
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("marshal snapshot %s: %w", key, err)
	}
	var out []T
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("unmarshal snapshot %s: %w", key, err)
	}
	return out, nil
}
