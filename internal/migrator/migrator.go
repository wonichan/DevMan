package migrator

import (
	"devman/internal/models"
	"devman/internal/registry"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type Engine struct {
	reg *registry.Registry
}

func New(reg *registry.Registry) *Engine {
	return &Engine{reg: reg}
}

type MigrationResult struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	BytesMoved int64  `json:"bytesMoved"`
	DurationMs int64  `json:"durationMs"`
}

type migrationRun struct {
	engine      *Engine
	envID       int64
	targetDir   string
	useJunction bool
	progress    func(models.MigrationProgress)
	startedAt   time.Time

	result      *MigrationResult
	snapshot    *models.Snapshot
	paths       []models.EnvPath
	installPath string
	stagingDir  string
	finalDir    string
	bytesMoved  int64
}

type migrationStep struct {
	name    string
	message string
	run     func(*migrationRun) error
}

func (e *Engine) Migrate(envID int64, targetDir string, useJunction bool) (*MigrationResult, error) {
	return e.MigrateWithProgress(envID, targetDir, useJunction, nil)
}

func (e *Engine) MigrateWithProgress(envID int64, targetDir string, useJunction bool, progress func(models.MigrationProgress)) (*MigrationResult, error) {
	run := &migrationRun{
		engine:      e,
		envID:       envID,
		targetDir:   targetDir,
		useJunction: useJunction,
		progress:    progress,
		startedAt:   time.Now(),
		result:      &MigrationResult{Success: false},
	}
	return run.execute()
}

func (r *migrationRun) execute() (*MigrationResult, error) {
	steps := []migrationStep{
		{name: "precheck", message: "Checking prerequisites", run: (*migrationRun).precheck},
		{name: "snapshot", message: "Creating snapshot", run: (*migrationRun).createSnapshot},
		{name: "paths", message: "Reading paths", run: (*migrationRun).readPaths},
		{name: "staging", message: "Creating staging directory", run: (*migrationRun).createStagingDir},
		{name: "copy", message: "Copying files", run: (*migrationRun).copyFiles},
		{name: "verify", message: "Verifying copy", run: (*migrationRun).verifyFiles},
		{name: "commit", message: "Committing migration", run: (*migrationRun).commitStaging},
		{name: "envvars", message: "Updating environment variables", run: (*migrationRun).updateEnvironment},
		{name: "junction", message: "Creating junction or removing source", run: (*migrationRun).finalizeSourcePath},
		{name: "registry", message: "Updating registry", run: (*migrationRun).updateRegistry},
	}

	logrus.WithFields(logrus.Fields{"env_id": r.envID, "target_dir": r.targetDir, "use_junction": r.useJunction}).Info("migration pipeline started")
	for index, step := range steps {
		r.emit(index, step.name, step.message, len(steps))
		if err := step.run(r); err != nil {
			r.result.Message = err.Error()
			logrus.WithError(err).WithFields(logrus.Fields{"env_id": r.envID, "step": step.name}).Warn("migration step failed")
			return r.result, nil
		}
	}

	r.result.Success = true
	r.result.Message = fmt.Sprintf("migration complete: %s -> %s", r.installPath, r.finalDir)
	r.result.BytesMoved = r.bytesMoved
	r.result.DurationMs = time.Since(r.startedAt).Milliseconds()
	r.saveHistory()

	r.emit(len(steps), "complete", "Migration complete", len(steps))
	logrus.WithFields(logrus.Fields{"env_id": r.envID, "bytes_moved": r.bytesMoved, "duration_ms": r.result.DurationMs}).Info("migration pipeline completed")
	return r.result, nil
}

func (r *migrationRun) emit(stepIndex int, step, message string, totalSteps int) {
	logrus.WithFields(logrus.Fields{"env_id": r.envID, "step": step, "step_index": stepIndex, "message": message}).Info("migration step")
	if r.progress == nil {
		return
	}
	percent := 100
	if stepIndex < totalSteps {
		percent = stepIndex * 100 / totalSteps
	}
	r.progress(models.MigrationProgress{
		Step:       step,
		StepIndex:  stepIndex,
		TotalSteps: totalSteps,
		Percent:    percent,
		Message:    message,
		EnvKey:     fmt.Sprintf("env_%d", r.envID),
	})
}

func (r *migrationRun) precheck() error {
	return r.engine.preCheck(r.envID, r.targetDir)
}

func (r *migrationRun) createSnapshot() error {
	snap, err := r.engine.createSnapshot(r.envID)
	if err != nil {
		return fmt.Errorf("create snapshot failed: %w", err)
	}
	r.snapshot = snap
	return nil
}

func (r *migrationRun) readPaths() error {
	paths, err := r.engine.reg.ListPaths(r.envID)
	if err != nil {
		return fmt.Errorf("read paths failed: %w", err)
	}
	for _, p := range paths {
		if p.Type == models.PathInstall {
			r.installPath = p.Path
			break
		}
	}
	if r.installPath == "" {
		return fmt.Errorf("install path not found")
	}
	r.paths = paths
	logrus.WithFields(logrus.Fields{"env_id": r.envID, "install_path": r.installPath, "path_count": len(paths)}).Info("migration paths loaded")
	return nil
}

func (r *migrationRun) createStagingDir() error {
	baseName := filepath.Base(r.installPath)
	r.stagingDir = filepath.Join(r.targetDir, ".devman_tmp", fmt.Sprintf("%s_%d", baseName, time.Now().Unix()))
	if err := os.MkdirAll(r.stagingDir, 0755); err != nil {
		return fmt.Errorf("create staging directory failed: %w", err)
	}
	logrus.WithField("staging_dir", r.stagingDir).Info("migration staging directory created")
	return nil
}

func (r *migrationRun) copyFiles() error {
	bytesMoved, err := r.engine.copyDir(r.installPath, r.stagingDir)
	if err != nil {
		_ = os.RemoveAll(r.stagingDir)
		return fmt.Errorf("copy failed: %w", err)
	}
	r.bytesMoved = bytesMoved
	logrus.WithFields(logrus.Fields{"source": r.installPath, "staging_dir": r.stagingDir, "bytes_moved": bytesMoved}).Info("migration copy completed")
	return nil
}

func (r *migrationRun) verifyFiles() error {
	if r.engine.verifyCopy(r.installPath, r.stagingDir) {
		logrus.WithFields(logrus.Fields{"source": r.installPath, "staging_dir": r.stagingDir}).Info("migration copy verified")
		return nil
	}
	_ = os.RemoveAll(r.stagingDir)
	return fmt.Errorf("copy verification failed")
}

func (r *migrationRun) commitStaging() error {
	baseName := filepath.Base(r.installPath)
	r.finalDir = filepath.Join(r.targetDir, baseName)

	if _, err := os.Stat(r.finalDir); err == nil {
		_ = os.RemoveAll(r.stagingDir)
		return fmt.Errorf("target directory already exists: %s", r.finalDir)
	} else if !os.IsNotExist(err) {
		_ = os.RemoveAll(r.stagingDir)
		return fmt.Errorf("failed to check target directory: %w", err)
	}

	if err := os.Rename(r.stagingDir, r.finalDir); err != nil {
		_ = r.engine.restoreSnapshot(r.snapshot)
		_ = os.RemoveAll(r.stagingDir)
		return fmt.Errorf("commit failed: %w", err)
	}
	logrus.WithFields(logrus.Fields{"staging_dir": r.stagingDir, "final_dir": r.finalDir}).Info("migration committed")
	return nil
}

func (r *migrationRun) updateEnvironment() error {
	if err := r.engine.updateEnvVars(r.installPath, r.finalDir); err != nil {
		_ = os.RemoveAll(r.finalDir)
		return fmt.Errorf("failed to update environment variables: %w", err)
	}
	return nil
}

func (r *migrationRun) finalizeSourcePath() error {
	if r.useJunction && runtime.GOOS == "windows" {
		if err := os.RemoveAll(r.installPath); err != nil {
			return fmt.Errorf("failed to remove source before creating junction: %w", err)
		}
		if err := r.engine.createJunction(r.installPath, r.finalDir); err != nil {
			return fmt.Errorf("failed to create junction: %w", err)
		}
		logrus.WithFields(logrus.Fields{"old_path": r.installPath, "new_path": r.finalDir}).Info("migration junction created")
		return nil
	}
	if err := os.RemoveAll(r.installPath); err != nil {
		logrus.WithError(err).WithField("path", r.installPath).Warn("migration failed to remove source")
	}
	return nil
}

func (r *migrationRun) updateRegistry() error {
	for i := range r.paths {
		if r.paths[i].Type == models.PathInstall {
			r.paths[i].Path = r.finalDir
		}
		if err := r.engine.reg.SavePath(&r.paths[i]); err != nil {
			logrus.WithError(err).WithField("path", r.paths[i].Path).Warn("migration failed to update registry path")
		}
	}
	return nil
}

func (r *migrationRun) saveHistory() {
	if err := r.engine.reg.SaveHistory(&models.HistoryEntry{
		Action:      "migrate",
		TargetEnv:   fmt.Sprintf("env_%d", r.envID),
		DetailsJSON: fmt.Sprintf(`{"from":"%s","to":"%s","bytes":%d}`, r.installPath, r.finalDir, r.bytesMoved),
		Success:     true,
		CreatedAt:   time.Now(),
	}); err != nil {
		logrus.WithError(err).WithField("env_id", r.envID).Warn("migration failed to save history")
	}
}

func (e *Engine) preCheck(envID int64, targetDir string) error {
	paths, err := e.reg.ListPaths(envID)
	if err != nil {
		logrus.WithError(err).WithField("env_id", envID).Error("migration precheck failed to list paths")
		return err
	}

	var totalSize int64
	installPath := ""
	for _, p := range paths {
		totalSize += p.SizeBytes
		if p.Type == models.PathInstall {
			installPath = p.Path
		}
	}
	if installPath == "" {
		logrus.WithField("env_id", envID).Warn("migration precheck missing install path")
		return fmt.Errorf("install path not found")
	}
	if err := validateMigrationPaths(installPath, targetDir); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"source": installPath, "target_dir": targetDir}).Warn("migration path validation failed")
		return err
	}
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		logrus.WithError(err).WithField("target_dir", targetDir).Error("migration target directory is not writable")
		return fmt.Errorf("target directory is not writable: %w", err)
	}
	if totalSize == 0 {
		logrus.WithField("env_id", envID).Warn("migration precheck source size is zero")
		return fmt.Errorf("source size is zero")
	}

	logrus.WithFields(logrus.Fields{"env_id": envID, "install_path": installPath, "target_dir": targetDir, "total_size": totalSize}).Info("migration precheck passed")
	return nil
}

func validateMigrationPaths(sourcePath, targetDir string) error {
	sourceAbs, err := filepath.Abs(sourcePath)
	if err != nil {
		return fmt.Errorf("invalid source path: %w", err)
	}
	targetAbs, err := filepath.Abs(targetDir)
	if err != nil {
		return fmt.Errorf("invalid target path: %w", err)
	}
	sourceAbs = filepath.Clean(sourceAbs)
	targetAbs = filepath.Clean(targetAbs)

	info, err := os.Stat(sourceAbs)
	if err != nil {
		return fmt.Errorf("source path is not accessible: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("source path is not a directory")
	}
	if isDangerousMigrationSource(sourceAbs) {
		return fmt.Errorf("refusing to migrate broad system directory: %s", sourceAbs)
	}
	if isDangerousMigrationTarget(targetAbs) {
		return fmt.Errorf("refusing to use broad target directory: %s", targetAbs)
	}
	if samePath(sourceAbs, targetAbs) || isPathInside(targetAbs, sourceAbs) {
		return fmt.Errorf("target directory cannot be inside source directory")
	}
	return nil
}

func isDangerousMigrationSource(path string) bool {
	clean := filepath.Clean(path)
	if filepath.Dir(clean) == clean {
		return true
	}
	volume := filepath.VolumeName(clean)
	if volume != "" && samePath(clean, volume+string(os.PathSeparator)) {
		return true
	}
	for _, base := range []string{userHomeDir(), os.TempDir()} {
		if base != "" && samePath(clean, base) {
			return true
		}
	}
	return false
}

func isDangerousMigrationTarget(path string) bool {
	clean := filepath.Clean(path)
	if filepath.Dir(clean) == clean {
		return true
	}
	volume := filepath.VolumeName(clean)
	if volume != "" && samePath(clean, volume+string(os.PathSeparator)) {
		return true
	}
	for _, base := range []string{userHomeDir(), os.TempDir()} {
		if base != "" && samePath(clean, base) {
			return true
		}
	}
	return false
}

func userHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Clean(home)
}

func samePath(a, b string) bool {
	return comparablePath(a) == comparablePath(b)
}

func isPathInside(path, parent string) bool {
	cleanPath := comparablePath(path)
	cleanParent := comparablePath(parent)
	if strings.HasPrefix(cleanPath, cleanParent+string(os.PathSeparator)) {
		return true
	}
	rel, err := filepath.Rel(parent, path)
	if err != nil {
		return false
	}
	return rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

func comparablePath(path string) string {
	clean := filepath.Clean(path)
	if runtime.GOOS == "windows" {
		return strings.ToLower(clean)
	}
	return clean
}

func (e *Engine) createSnapshot(envID int64) (*models.Snapshot, error) {
	data, err := e.reg.ExportSnapshotData()
	if err != nil {
		logrus.WithError(err).WithField("env_id", envID).Error("failed to export snapshot data")
		return nil, err
	}
	b, err := json.Marshal(data)
	if err != nil {
		logrus.WithError(err).WithField("env_id", envID).Error("failed to marshal snapshot data")
		return nil, err
	}
	snap := &models.Snapshot{
		Name:      fmt.Sprintf("pre-migrate-%d-%d", envID, time.Now().Unix()),
		DataJSON:  string(b),
		CreatedAt: time.Now(),
	}
	if err := e.reg.SaveSnapshot(snap); err != nil {
		logrus.WithError(err).WithField("env_id", envID).Error("failed to save migration snapshot")
		return nil, err
	}
	logrus.WithFields(logrus.Fields{"env_id": envID, "snapshot_id": snap.ID}).Info("migration snapshot created")
	return snap, nil
}

func (e *Engine) restoreSnapshot(snap *models.Snapshot) error {
	if snap == nil {
		return nil
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(snap.DataJSON), &data); err != nil {
		logrus.WithError(err).WithField("snapshot_id", snap.ID).Error("failed to unmarshal snapshot data")
		return err
	}
	if err := e.reg.ImportSnapshotData(data); err != nil {
		logrus.WithError(err).WithField("snapshot_id", snap.ID).Error("failed to restore snapshot")
		return err
	}
	logrus.WithField("snapshot_id", snap.ID).Info("snapshot restored")
	return nil
}

func (e *Engine) copyDir(src, dst string) (int64, error) {
	var total int64
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logrus.WithError(err).WithField("path", path).Error("migration copy walk failed")
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, rel)
		if info.IsDir() {
			if err := os.MkdirAll(dstPath, info.Mode()); err != nil {
				logrus.WithError(err).WithField("path", dstPath).Error("migration failed to create destination directory")
				return err
			}
			return nil
		}
		if err := copyFile(path, dstPath); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{"source": path, "destination": dstPath}).Error("migration failed to copy file")
			return err
		}
		total += info.Size()
		return nil
	})
	return total, err
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func (e *Engine) verifyCopy(src, dst string) bool {
	var srcSize, dstSize int64
	var srcCount, dstCount int

	filepath.Walk(src, func(_ string, info os.FileInfo, _ error) error {
		if info != nil && !info.IsDir() {
			srcSize += info.Size()
			srcCount++
		}
		return nil
	})
	filepath.Walk(dst, func(_ string, info os.FileInfo, _ error) error {
		if info != nil && !info.IsDir() {
			dstSize += info.Size()
			dstCount++
		}
		return nil
	})

	matched := srcCount == dstCount && srcSize == dstSize
	logrus.WithFields(logrus.Fields{"source": src, "destination": dst, "source_files": srcCount, "destination_files": dstCount, "source_size": srcSize, "destination_size": dstSize, "matched": matched}).Info("migration copy verification result")
	return matched
}

func (e *Engine) updateEnvVars(oldPath, newPath string) error {
	if runtime.GOOS == "windows" {
		return updateWindowsPath(oldPath, newPath)
	}
	return nil
}

func (e *Engine) createJunction(oldPath, newPath string) error {
	if runtime.GOOS != "windows" {
		return nil
	}
	return createWindowsJunction(oldPath, newPath)
}
