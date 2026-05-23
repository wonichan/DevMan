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

func (e *Engine) Migrate(envID int64, targetDir string, useJunction bool) (*MigrationResult, error) {
	return e.MigrateWithProgress(envID, targetDir, useJunction, nil)
}

func (e *Engine) MigrateWithProgress(envID int64, targetDir string, useJunction bool, progress func(models.MigrationProgress)) (*MigrationResult, error) {
	start := time.Now()
	logrus.WithFields(logrus.Fields{"env_id": envID, "target_dir": targetDir, "use_junction": useJunction}).Info("migration pipeline started")
	result := &MigrationResult{Success: false}
	steps := []string{"precheck", "snapshot", "paths", "staging", "copy", "verify", "commit", "envvars", "junction", "registry"}
	emit := func(stepIndex int, step, message string) {
		logrus.WithFields(logrus.Fields{"env_id": envID, "step": step, "step_index": stepIndex, "message": message}).Info("migration step")
		if progress == nil {
			return
		}
		percent := 100
		if stepIndex < len(steps) {
			percent = stepIndex * 100 / len(steps)
		}
		progress(models.MigrationProgress{
			Step:       step,
			StepIndex:  stepIndex,
			TotalSteps: len(steps),
			Percent:    percent,
			Message:    message,
			EnvKey:     fmt.Sprintf("env_%d", envID),
		})
	}

	// 1. Pre-check
	emit(0, "precheck", "Checking prerequisites")
	if err := e.preCheck(envID, targetDir); err != nil {
		result.Message = err.Error()
		logrus.WithError(err).WithField("env_id", envID).Warn("migration precheck failed")
		return result, nil
	}

	// 2. Snapshot
	emit(1, "snapshot", "Creating snapshot")
	snap, err := e.createSnapshot(envID)
	if err != nil {
		result.Message = "创建快照失败: " + err.Error()
		logrus.WithError(err).WithField("env_id", envID).Error("migration snapshot failed")
		return result, nil
	}

	// 3. Get paths to migrate
	emit(2, "paths", "Reading paths")
	paths, err := e.reg.ListPaths(envID)
	if err != nil {
		result.Message = "读取路径失败"
		logrus.WithError(err).WithField("env_id", envID).Error("migration failed to read paths")
		return result, nil
	}

	var installPath string
	for _, p := range paths {
		if p.Type == models.PathInstall {
			installPath = p.Path
			break
		}
	}
	if installPath == "" {
		result.Message = "未找到安装路径"
		logrus.WithField("env_id", envID).Warn("migration install path not found")
		return result, nil
	}
	logrus.WithFields(logrus.Fields{"env_id": envID, "install_path": installPath, "path_count": len(paths)}).Info("migration paths loaded")

	// 4. Create staging dir
	emit(3, "staging", "Creating staging directory")
	baseName := filepath.Base(installPath)
	stagingDir := filepath.Join(targetDir, ".devman_tmp", fmt.Sprintf("%s_%d", baseName, time.Now().Unix()))
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		result.Message = "创建临时目录失败: " + err.Error()
		logrus.WithError(err).WithField("staging_dir", stagingDir).Error("migration failed to create staging directory")
		return result, nil
	}
	logrus.WithField("staging_dir", stagingDir).Info("migration staging directory created")

	// 5. Copy files
	emit(4, "copy", "Copying files")
	bytesMoved, err := e.copyDir(installPath, stagingDir)
	if err != nil {
		// Rollback: remove staging
		_ = os.RemoveAll(stagingDir)
		result.Message = "复制失败: " + err.Error()
		logrus.WithError(err).WithFields(logrus.Fields{"source": installPath, "staging_dir": stagingDir}).Error("migration copy failed")
		return result, nil
	}
	logrus.WithFields(logrus.Fields{"source": installPath, "staging_dir": stagingDir, "bytes_moved": bytesMoved}).Info("migration copy completed")

	// 6. Verify
	emit(5, "verify", "Verifying copy")
	if !e.verifyCopy(installPath, stagingDir) {
		_ = os.RemoveAll(stagingDir)
		result.Message = "验证失败，移除临时目录"
		logrus.WithFields(logrus.Fields{"source": installPath, "staging_dir": stagingDir}).Error("migration copy verification failed")
		return result, nil
	}
	logrus.WithFields(logrus.Fields{"source": installPath, "staging_dir": stagingDir}).Info("migration copy verified")

	// Commit file copy before mutating environment variables.
	finalDir := filepath.Join(targetDir, baseName)
	if _, err := os.Stat(finalDir); err == nil {
		_ = os.RemoveAll(stagingDir)
		result.Message = "target directory already exists: " + finalDir
		logrus.WithField("final_dir", finalDir).Warn("migration target final directory already exists")
		return result, nil
	} else if !os.IsNotExist(err) {
		_ = os.RemoveAll(stagingDir)
		result.Message = "failed to check target directory: " + err.Error()
		logrus.WithError(err).WithField("final_dir", finalDir).Error("migration failed to check final directory")
		return result, nil
	}
	// 7. Commit: rename staging to final.
	emit(6, "commit", "Committing migration")
	if err := os.Rename(stagingDir, finalDir); err != nil {
		_ = e.restoreSnapshot(snap)
		_ = os.RemoveAll(stagingDir)
		result.Message = "提交失败: " + err.Error()
		logrus.WithError(err).WithFields(logrus.Fields{"staging_dir": stagingDir, "final_dir": finalDir}).Error("migration commit failed")
		return result, nil
	}
	logrus.WithFields(logrus.Fields{"staging_dir": stagingDir, "final_dir": finalDir}).Info("migration committed")

	// 8. Update environment variables after the copy has been committed.
	emit(7, "envvars", "Updating environment variables")
	if err := e.updateEnvVars(installPath, finalDir); err != nil {
		_ = os.RemoveAll(finalDir)
		result.Message = "failed to update environment variables: " + err.Error()
		logrus.WithError(err).WithFields(logrus.Fields{"old_path": installPath, "new_path": finalDir}).Error("migration environment variable update failed")
		return result, nil
	}

	// 9. Create junction if requested (Windows only)
	emit(8, "junction", "Creating junction or removing source")
	if useJunction && runtime.GOOS == "windows" {
		if err := os.RemoveAll(installPath); err != nil {
			result.Message = "failed to remove source before creating junction: " + err.Error()
			logrus.WithError(err).WithField("path", installPath).Error("migration failed to remove source before junction")
			return result, nil
		}
		if err := e.createJunction(installPath, finalDir); err != nil {
			result.Message = "failed to create junction: " + err.Error()
			logrus.WithError(err).WithFields(logrus.Fields{"old_path": installPath, "new_path": finalDir}).Error("migration junction creation failed")
			return result, nil
		} else {
			logrus.WithFields(logrus.Fields{"old_path": installPath, "new_path": finalDir}).Info("migration junction created")
		}
	} else {
		// Remove source
		if err := os.RemoveAll(installPath); err != nil {
			logrus.WithError(err).WithField("path", installPath).Warn("migration failed to remove source")
		}
	}

	// 10. Update registry
	emit(9, "registry", "Updating registry")
	for i := range paths {
		if paths[i].Type == models.PathInstall {
			paths[i].Path = finalDir
		}
		if err := e.reg.SavePath(&paths[i]); err != nil {
			logrus.WithError(err).WithField("path", paths[i].Path).Warn("migration failed to update registry path")
		}
	}

	result.Success = true
	result.Message = fmt.Sprintf("迁移成功: %s → %s", installPath, finalDir)
	result.BytesMoved = bytesMoved
	result.DurationMs = time.Since(start).Milliseconds()

	if err := e.reg.SaveHistory(&models.HistoryEntry{
		Action:      "migrate",
		TargetEnv:   fmt.Sprintf("env_%d", envID),
		DetailsJSON: fmt.Sprintf(`{"from":"%s","to":"%s","bytes":%d}`, installPath, finalDir, bytesMoved),
		Success:     true,
		CreatedAt:   time.Now(),
	}); err != nil {
		logrus.WithError(err).WithField("env_id", envID).Warn("migration failed to save history")
	}
	emit(len(steps), "complete", "Migration complete")
	logrus.WithFields(logrus.Fields{"env_id": envID, "bytes_moved": bytesMoved, "duration_ms": result.DurationMs}).Info("migration pipeline completed")

	return result, nil
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
		return fmt.Errorf("未找到安装路径")
	}
	if err := validateMigrationPaths(installPath, targetDir); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"source": installPath, "target_dir": targetDir}).Warn("migration path validation failed")
		return err
	}
	// Check target dir exists or can be created after validating path safety.
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		logrus.WithError(err).WithField("target_dir", targetDir).Error("migration target directory is not writable")
		return fmt.Errorf("目标目录不可写: %w", err)
	}
	// TODO: check actual free space on target disk
	if totalSize == 0 {
		logrus.WithField("env_id", envID).Warn("migration precheck source size is zero")
		return fmt.Errorf("未能计算源目录大小")
	}
	logrus.WithFields(logrus.Fields{"env_id": envID, "install_path": installPath, "target_dir": targetDir, "total_size": totalSize}).Info("migration precheck passed")
	return nil
}

func validateMigrationPaths(sourcePath, targetDir string) error {
	sourceAbs, err := filepath.Abs(sourcePath)
	if err != nil {
		return fmt.Errorf("源路径无效: %w", err)
	}
	targetAbs, err := filepath.Abs(targetDir)
	if err != nil {
		return fmt.Errorf("目标路径无效: %w", err)
	}
	sourceAbs = filepath.Clean(sourceAbs)
	targetAbs = filepath.Clean(targetAbs)

	info, err := os.Stat(sourceAbs)
	if err != nil {
		return fmt.Errorf("源路径不可访问: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("源路径不是目录")
	}
	if isDangerousMigrationSource(sourceAbs) {
		return fmt.Errorf("拒绝迁移过宽的系统目录: %s", sourceAbs)
	}
	if isDangerousMigrationTarget(targetAbs) {
		return fmt.Errorf("拒绝使用过宽的目标目录: %s", targetAbs)
	}
	if samePath(sourceAbs, targetAbs) || isPathInside(targetAbs, sourceAbs) {
		return fmt.Errorf("目标目录不能位于源目录内部")
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
	// Simple verification: compare top-level file count and total size
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
	// Platform-specific: Windows registry PATH update
	if runtime.GOOS == "windows" {
		return updateWindowsPath(oldPath, newPath)
	}
	// Linux: update ~/.bashrc or similar (simplified)
	return nil
}

func (e *Engine) createJunction(oldPath, newPath string) error {
	// Windows-specific: create directory junction
	if runtime.GOOS != "windows" {
		return nil
	}
	return createWindowsJunction(oldPath, newPath)
}
