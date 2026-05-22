package registry

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"devman/internal/models"

	_ "github.com/mattn/go-sqlite3"
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
	db, err := sql.Open("sqlite3", dbPath())
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	r := &Registry{db: db}
	if err := r.migrate(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Registry) Close() error {
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
`
	_, err := r.db.Exec(schema)
	return err
}

func (r *Registry) SaveEnv(env *models.Env) error {
	res, err := r.db.Exec(
		`INSERT INTO envs (name, key, category, icon, description, website, is_managed, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(key) DO UPDATE SET
		   name=excluded.name, category=excluded.category, icon=excluded.icon,
		   description=excluded.description, website=excluded.website,
		   is_managed=excluded.is_managed, updated_at=excluded.updated_at`,
		env.Name, env.Key, env.Category, env.Icon, env.Description, env.Website, env.IsManaged,
		env.CreatedAt, env.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if env.ID == 0 {
		id, _ := res.LastInsertId()
		env.ID = id
	}
	return nil
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

func (r *Registry) ListInstances(envID int64) ([]models.EnvInstance, error) {
	rows, err := r.db.Query(`SELECT id, env_id, version, install_path, is_default, is_active, source, detected_at FROM env_instances WHERE env_id = ?`, envID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.EnvInstance
	for rows.Next() {
		var i models.EnvInstance
		var id, eid, def, act int
		if err := rows.Scan(&i.ID, &i.EnvID, &i.Version, &i.InstallPath, &def, &act, &i.Source, &i.DetectedAt); err != nil {
			return nil, err
		}
		_ = id
		_ = eid
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
		"envs":      envs,
		"instances": instList,
		"paths":     pathList,
		"exportedAt": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (r *Registry) ImportSnapshotData(data map[string]interface{}) error {
	// Clear existing
	_, _ = r.db.Exec(`DELETE FROM env_paths`)
	_, _ = r.db.Exec(`DELETE FROM env_instances`)
	_, _ = r.db.Exec(`DELETE FROM envs`)

	// Re-insert envs
	envsRaw, ok := data["envs"].([]interface{})
	if ok {
		for _, eRaw := range envsRaw {
			b, _ := json.Marshal(eRaw)
			var e models.Env
			json.Unmarshal(b, &e)
			if e.Key != "" {
				r.SaveEnv(&e)
			}
		}
	}
	// Note: instances and paths would need similar handling
	return nil
}
