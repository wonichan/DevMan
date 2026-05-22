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
	"time"
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
	start := time.Now()
	result := &MigrationResult{Success: false}

	// 1. Pre-check
	if err := e.preCheck(envID, targetDir); err != nil {
		result.Message = err.Error()
		return result, nil
	}

	// 2. Snapshot
	snap, err := e.createSnapshot(envID)
	if err != nil {
		result.Message = "创建快照失败: " + err.Error()
		return result, nil
	}

	// 3. Get paths to migrate
	paths, err := e.reg.ListPaths(envID)
	if err != nil {
		result.Message = "读取路径失败"
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
		return result, nil
	}

	// 4. Create staging dir
	baseName := filepath.Base(installPath)
	stagingDir := filepath.Join(targetDir, ".devman_tmp", fmt.Sprintf("%s_%d", baseName, time.Now().Unix()))
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		result.Message = "创建临时目录失败: " + err.Error()
		return result, nil
	}

	// 5. Copy files
	bytesMoved, err := e.copyDir(installPath, stagingDir)
	if err != nil {
		// Rollback: remove staging
		_ = os.RemoveAll(stagingDir)
		result.Message = "复制失败: " + err.Error()
		return result, nil
	}

	// 6. Verify
	if !e.verifyCopy(installPath, stagingDir) {
		_ = os.RemoveAll(stagingDir)
		result.Message = "验证失败，移除临时目录"
		return result, nil
	}

	// 7. Update environment variables (platform specific)
	finalDir := filepath.Join(targetDir, baseName)
	if err := e.updateEnvVars(installPath, finalDir); err != nil {
		_ = os.RemoveAll(stagingDir)
		result.Message = "更新环境变量失败: " + err.Error()
		return result, nil
	}

	// 8. Commit: rename staging to final
	if err := os.Rename(stagingDir, finalDir); err != nil {
		// Rollback env vars
		_ = e.restoreSnapshot(snap)
		_ = os.RemoveAll(stagingDir)
		result.Message = "提交失败: " + err.Error()
		return result, nil
	}

	// 9. Create junction if requested (Windows only)
	if useJunction && runtime.GOOS == "windows" {
		_ = e.createJunction(installPath, finalDir)
	} else {
		// Remove source
		_ = os.RemoveAll(installPath)
	}

	// 10. Update registry
	for i := range paths {
		if paths[i].Type == models.PathInstall {
			paths[i].Path = finalDir
		}
		_ = e.reg.SavePath(&paths[i])
	}

	result.Success = true
	result.Message = fmt.Sprintf("迁移成功: %s → %s", installPath, finalDir)
	result.BytesMoved = bytesMoved
	result.DurationMs = time.Since(start).Milliseconds()

	_ = e.reg.SaveHistory(&models.HistoryEntry{
		Action:      "migrate",
		TargetEnv:   fmt.Sprintf("env_%d", envID),
		DetailsJSON: fmt.Sprintf(`{"from":"%s","to":"%s","bytes":%d}`, installPath, finalDir, bytesMoved),
		Success:     true,
		CreatedAt:   time.Now(),
	})

	return result, nil
}

func (e *Engine) preCheck(envID int64, targetDir string) error {
	// Check target dir exists or can be created
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("目标目录不可写: %w", err)
	}
	// Check target disk space (simplified)
	paths, _ := e.reg.ListPaths(envID)
	var totalSize int64
	for _, p := range paths {
		totalSize += p.SizeBytes
	}
	// TODO: check actual free space on target disk
	if totalSize == 0 {
		return fmt.Errorf("未能计算源目录大小")
	}
	return nil
}

func (e *Engine) createSnapshot(envID int64) (*models.Snapshot, error) {
	data, err := e.reg.ExportSnapshotData()
	if err != nil {
		return nil, err
	}
	b, _ := json.Marshal(data)
	snap := &models.Snapshot{
		Name:     fmt.Sprintf("pre-migrate-%d-%d", envID, time.Now().Unix()),
		DataJSON: string(b),
		CreatedAt: time.Now(),
	}
	if err := e.reg.SaveSnapshot(snap); err != nil {
		return nil, err
	}
	return snap, nil
}

func (e *Engine) restoreSnapshot(snap *models.Snapshot) error {
	if snap == nil {
		return nil
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(snap.DataJSON), &data); err != nil {
		return err
	}
	return e.reg.ImportSnapshotData(data)
}

func (e *Engine) copyDir(src, dst string) (int64, error) {
	var total int64
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		if err := copyFile(path, dstPath); err != nil {
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

	return srcCount == dstCount && srcSize == dstSize
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
