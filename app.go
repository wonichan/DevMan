package main

import (
	"context"
	"devman/internal/migrator"
	"devman/internal/models"
	"devman/internal/registry"
	"devman/internal/scanner"
	"devman/internal/utils"
	"devman/internal/versionmanager"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx            context.Context
	reg            *registry.Registry
	engine         *scanner.Engine
	migrator       *migrator.Engine
	versionManager *versionmanager.Service
	scanMu         sync.Mutex
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	logrus.Info("application startup begin")
	a.ctx = ctx

	// Ensure db directory exists (portable mode: same dir as exe)
	if exe, err := os.Executable(); err == nil {
		dbDir := filepath.Dir(exe)
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			logrus.WithError(err).WithField("dir", dbDir).Error("failed to ensure executable directory")
		} else {
			logrus.WithField("dir", dbDir).Info("executable directory ready")
		}
	} else {
		logrus.WithError(err).Warn("failed to resolve executable path")
	}

	reg, err := registry.Open()
	if err != nil {
		logrus.WithError(err).Error("failed to open registry")
		return
	}
	a.reg = reg
	a.engine = scanner.NewEngine(reg)
	a.migrator = migrator.New(reg)
	a.versionManager = versionmanager.NewService(reg, versionmanager.RealEnvironment{})
	logrus.Info("application startup complete")
}

func (a *App) shutdown(ctx context.Context) {
	logrus.Info("application shutdown begin")
	if a.reg != nil {
		if err := a.reg.Close(); err != nil {
			logrus.WithError(err).Error("failed to close registry")
		} else {
			logrus.Info("registry closed")
		}
	}
	logrus.Info("application shutdown complete")
}

// ScanAll runs all scanners and returns summaries
func (a *App) ScanAll() ([]models.EnvSummary, error) {
	logrus.Info("scan all requested")
	if a.engine == nil {
		logrus.Error("scan all failed: engine not initialized")
		return nil, fmt.Errorf("engine not initialized")
	}
	if !a.scanMu.TryLock() {
		logrus.Warn("scan all skipped: scan already in progress")
		return nil, fmt.Errorf("scan already in progress")
	}
	defer a.scanMu.Unlock()
	settings, err := a.GetSettings()
	if err != nil {
		logrus.WithError(err).Error("scan all failed to load settings")
		return nil, err
	}
	summaries, err := a.engine.ScanAllWithOptions(scanner.ScanOptions{CustomScanPaths: settings.CustomScanPaths})
	if err != nil {
		logrus.WithError(err).Error("scan all failed")
		return nil, err
	}
	logrus.WithField("summary_count", len(summaries)).Info("scan all completed")
	a.saveScanMetricSnapshots(summaries)
	return summaries, nil
}

func (a *App) saveScanMetricSnapshots(summaries []models.EnvSummary) {
	if a.reg == nil {
		return
	}
	now := time.Now()
	var totalSize int64
	for _, s := range summaries {
		if s.TotalSize > 0 {
			if err := a.reg.SaveMetricSnapshot(&models.MetricSnapshot{
				MetricKey:  "env_total_size",
				TargetKey:  s.Env.Key,
				ValueBytes: s.TotalSize,
				CapturedAt: now,
			}); err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{"metric_key": "env_total_size", "target_key": s.Env.Key}).Warn("failed to save per-env metric snapshot")
			}
			totalSize += s.TotalSize
		}
	}
	if totalSize > 0 {
		if err := a.reg.SaveMetricSnapshot(&models.MetricSnapshot{
			MetricKey:  "env_total_size",
			TargetKey:  "all",
			ValueBytes: totalSize,
			CapturedAt: now,
		}); err != nil {
			logrus.WithError(err).WithField("metric_key", "env_total_size").Warn("failed to save aggregate metric snapshot")
		}
	}
	if err := a.reg.PruneMetricSnapshots("env_total_size", 1000); err != nil {
		logrus.WithError(err).WithField("metric_key", "env_total_size").Warn("failed to prune metric snapshots")
	}
}

// GetSettings returns the current application settings
func (a *App) GetSettings() (*models.AppSettings, error) {
	if a.reg == nil {
		logrus.Error("get settings failed: registry not initialized")
		return nil, fmt.Errorf("registry not initialized")
	}
	settings, err := a.reg.GetSettings()
	if err != nil {
		logrus.WithError(err).Error("get settings failed")
		return nil, err
	}
	return settings, nil
}

// SaveSettings persists application settings
func (a *App) SaveSettings(settings models.AppSettings) error {
	logrus.WithField("custom_scan_paths", len(settings.CustomScanPaths)).Info("save settings requested")
	if a.reg == nil {
		logrus.Error("save settings failed: registry not initialized")
		return fmt.Errorf("registry not initialized")
	}
	if err := a.reg.SaveSettings(&settings); err != nil {
		logrus.WithError(err).Error("save settings failed")
		return err
	}
	logrus.Info("settings saved")
	return nil
}

// GetEnvs returns all stored environments
func (a *App) GetEnvs() ([]models.Env, error) {
	if a.reg == nil {
		logrus.Error("get environments failed: registry not initialized")
		return nil, fmt.Errorf("registry not initialized")
	}
	envs, err := a.reg.ListEnvs()
	if err != nil {
		logrus.WithError(err).Error("get environments failed")
		return nil, err
	}
	logrus.WithField("env_count", len(envs)).Info("get environments completed")
	return envs, nil
}

// ManageEnv marks a detected environment as user-approved for DevMan tracking.
func (a *App) ManageEnv(key string) (*models.Env, error) {
	return a.setEnvManaged(key, true)
}

// UnmanageEnv removes DevMan tracking approval from a detected environment.
func (a *App) UnmanageEnv(key string) (*models.Env, error) {
	return a.setEnvManaged(key, false)
}

func (a *App) setEnvManaged(key string, managed bool) (*models.Env, error) {
	action := "unmanage_env"
	if managed {
		action = "manage_env"
	}
	logrus.WithFields(logrus.Fields{"env_key": key, "managed": managed}).Info("environment management update requested")
	if a.reg == nil {
		logrus.WithField("env_key", key).Error("environment management update failed: registry not initialized")
		return nil, fmt.Errorf("registry not initialized")
	}
	if strings.TrimSpace(key) == "" {
		logrus.Error("environment management update failed: env key is required")
		return nil, fmt.Errorf("env key is required")
	}
	env, err := a.reg.SetEnvManaged(key, managed)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"env_key": key, "managed": managed}).Error("environment management update failed")
		return nil, err
	}
	if err := a.reg.SaveHistory(&models.HistoryEntry{
		Action:      action,
		TargetEnv:   key,
		DetailsJSON: fmt.Sprintf(`{"key":"%s","managed":%t}`, key, managed),
		Success:     true,
		CreatedAt:   time.Now(),
	}); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"env_key": key, "managed": managed}).Warn("failed to save environment management history")
	}
	logrus.WithFields(logrus.Fields{"env_key": key, "managed": managed}).Info("environment management update completed")
	return env, nil
}

// GetEnvSummary returns full summary for an env
func (a *App) GetEnvSummary(key string) (*models.EnvSummary, error) {
	logrus.WithField("env_key", key).Info("get environment summary requested")
	if a.reg == nil {
		logrus.Error("get environment summary failed: registry not initialized")
		return nil, fmt.Errorf("registry not initialized")
	}
	env, err := a.reg.GetEnvByKey(key)
	if err != nil || env == nil {
		logrus.WithError(err).WithField("env_key", key).Warn("environment not found")
		return nil, fmt.Errorf("env not found: %s", key)
	}
	instances, err := a.reg.ListInstances(env.ID)
	if err != nil {
		logrus.WithError(err).WithField("env_id", env.ID).Warn("failed to list environment instances")
	}
	paths, err := a.reg.ListPaths(env.ID)
	if err != nil {
		logrus.WithError(err).WithField("env_id", env.ID).Warn("failed to list environment paths")
	}

	totalSize := int64(0)
	for _, p := range paths {
		totalSize += p.SizeBytes
	}

	health := models.HealthHealthy
	if totalSize > 5*1024*1024*1024 {
		health = models.HealthWarning
	}

	summary := &models.EnvSummary{
		Env:       *env,
		Instances: instances,
		Paths:     paths,
		TotalSize: totalSize,
		Health:    health,
	}
	logrus.WithFields(logrus.Fields{"env_key": key, "instances": len(instances), "paths": len(paths), "total_size": totalSize}).Info("get environment summary completed")
	return summary, nil
}

// Migrate moves an env to target directory
func (a *App) Migrate(envID int64, targetDir string, useJunction bool) (*migrator.MigrationResult, error) {
	logrus.WithFields(logrus.Fields{"env_id": envID, "target_dir": targetDir, "use_junction": useJunction}).Info("migration requested")
	if a.migrator == nil {
		logrus.Error("migration failed: migrator not initialized")
		return nil, fmt.Errorf("migrator not initialized")
	}

	emitProgress := func(progress models.MigrationProgress) {
		if a.ctx != nil {
			wailsRuntime.EventsEmit(a.ctx, "migration:progress", progress)
		}
	}
	result, err := a.migrator.MigrateWithProgress(envID, targetDir, useJunction, emitProgress)
	if err != nil {
		logrus.WithError(err).WithField("env_id", envID).Error("migration failed with error")
		return nil, err
	}
	if result != nil && result.Success {
		logrus.WithFields(logrus.Fields{"env_id": envID, "bytes_moved": result.BytesMoved, "duration_ms": result.DurationMs}).Info("migration completed")
	} else if result != nil {
		logrus.WithFields(logrus.Fields{"env_id": envID, "message": result.Message}).Warn("migration completed without success")
	}
	return result, nil
}

// GetDiskInfo returns disk usage info (stub on Linux)
func (a *App) GetDiskInfo() ([]models.DiskInfo, error) {
	disks, err := utils.GetDiskInfo()
	if err != nil {
		logrus.WithError(err).Error("get disk info failed")
		return nil, err
	}
	logrus.WithField("disk_count", len(disks)).Info("get disk info completed")
	return disks, nil
}

// Platform-specific disk info removed — now in utils package
func (a *App) GetHistory(limit int) ([]models.HistoryEntry, error) {
	if a.reg == nil {
		logrus.Error("get history failed: registry not initialized")
		return nil, fmt.Errorf("registry not initialized")
	}
	history, err := a.reg.GetHistory(limit)
	if err != nil {
		logrus.WithError(err).WithField("limit", limit).Error("get history failed")
		return nil, err
	}
	logrus.WithFields(logrus.Fields{"limit": limit, "history_count": len(history)}).Info("get history completed")
	return history, nil
}

// GetMetricSnapshots returns metric snapshots for trend cards
func (a *App) GetMetricSnapshots(metricKey string, targetKey string, limit int) ([]models.MetricSnapshot, error) {
	if a.reg == nil {
		logrus.Error("get metric snapshots failed: registry not initialized")
		return nil, fmt.Errorf("registry not initialized")
	}
	if limit <= 0 {
		limit = 100
	}
	snapshots, err := a.reg.ListMetricSnapshots(metricKey, targetKey, limit)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"metric_key": metricKey, "target_key": targetKey}).Error("get metric snapshots failed")
		return nil, err
	}
	logrus.WithFields(logrus.Fields{"metric_key": metricKey, "target_key": targetKey, "count": len(snapshots)}).Info("get metric snapshots completed")
	return snapshots, nil
}

// ListToolVersions returns supported version-managed tools and known local versions.
func (a *App) ListToolVersions() ([]versionmanager.ToolVersionState, error) {
	if a.versionManager == nil {
		logrus.Error("list tool versions failed: version manager not initialized")
		return nil, fmt.Errorf("version manager not initialized")
	}
	states, err := a.versionManager.ListToolVersions()
	if err != nil {
		logrus.WithError(err).Error("list tool versions failed")
		return nil, err
	}
	logrus.WithField("tool_count", len(states)).Info("list tool versions completed")
	return states, nil
}

// PreviewVersionInstall resolves where a version install would be placed.
func (a *App) PreviewVersionInstall(toolKey string, version string) (*versionmanager.VersionInstallPlan, error) {
	logrus.WithFields(logrus.Fields{"tool_key": toolKey, "version": version}).Info("preview version install requested")
	if a.versionManager == nil {
		logrus.Error("preview version install failed: version manager not initialized")
		return nil, fmt.Errorf("version manager not initialized")
	}
	plan, err := a.versionManager.PreviewVersionInstall(toolKey, version)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"tool_key": toolKey, "version": version}).Warn("preview version install failed")
		return nil, err
	}
	return plan, nil
}

// InstallVersion downloads, extracts, and records a DevMan-managed tool version.
func (a *App) InstallVersion(toolKey string, version string, targetDir string) (*versionmanager.VersionOperationResult, error) {
	logrus.WithFields(logrus.Fields{"tool_key": toolKey, "version": version, "target_dir": targetDir}).Info("install version requested")
	if a.versionManager == nil {
		logrus.Error("install version failed: version manager not initialized")
		return nil, fmt.Errorf("version manager not initialized")
	}
	result, err := a.versionManager.InstallVersion(toolKey, version, targetDir)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"tool_key": toolKey, "version": version, "target_dir": targetDir}).Warn("install version failed")
		return nil, err
	}
	logrus.WithFields(logrus.Fields{"tool_key": toolKey, "version": version, "target_dir": targetDir}).Info("install version completed")
	return result, nil
}

// FetchOfficialVersions retrieves available versions from the tool's official source.
func (a *App) FetchOfficialVersions(toolKey string) (*versionmanager.ToolVersionCatalog, error) {
	logrus.WithField("tool_key", toolKey).Info("fetch official versions requested")
	if a.versionManager == nil {
		logrus.Error("fetch official versions failed: version manager not initialized")
		return nil, fmt.Errorf("version manager not initialized")
	}
	catalog, err := a.versionManager.FetchOfficialVersions(toolKey)
	if err != nil {
		logrus.WithError(err).WithField("tool_key", toolKey).Warn("fetch official versions failed")
		return nil, err
	}
	logrus.WithFields(logrus.Fields{"tool_key": toolKey, "version_count": len(catalog.Versions)}).Info("fetch official versions completed")
	return catalog, nil
}

// SwitchVersion makes a tracked tool version the active user version through DevMan shims.
func (a *App) SwitchVersion(toolKey string, instanceId int64) (*versionmanager.VersionOperationResult, error) {
	logrus.WithFields(logrus.Fields{"tool_key": toolKey, "instance_id": instanceId}).Info("switch version requested")
	if a.versionManager == nil {
		logrus.Error("switch version failed: version manager not initialized")
		return nil, fmt.Errorf("version manager not initialized")
	}
	if a.reg == nil {
		logrus.Error("switch version failed: registry not initialized")
		return nil, fmt.Errorf("registry not initialized")
	}

	versions, err := a.reg.ListToolVersions(toolKey)
	if err != nil {
		logrus.WithError(err).WithField("tool_key", toolKey).Error("switch version failed to list versions")
		return nil, err
	}
	for _, version := range versions {
		if version.ID == instanceId {
			result, err := a.versionManager.SwitchVersion(version)
			if err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{"tool_key": toolKey, "instance_id": instanceId}).Warn("switch version failed")
				return nil, err
			}
			logrus.WithFields(logrus.Fields{"tool_key": toolKey, "instance_id": instanceId}).Info("switch version completed")
			return result, nil
		}
	}
	logrus.WithFields(logrus.Fields{"tool_key": toolKey, "instance_id": instanceId}).Warn("switch version failed: version not found")
	return nil, fmt.Errorf("version not found: %d", instanceId)
}

// DetectVersionManager reports external version-manager ownership for a tool.
func (a *App) DetectVersionManager(toolKey string) *versionmanager.VersionManagerConflict {
	if a.versionManager == nil {
		logrus.WithField("tool_key", toolKey).Warn("detect version manager skipped: version manager not initialized")
		return nil
	}
	return a.versionManager.DetectVersionManager(toolKey)
}

// AnalyzeCleanable returns items that can be safely cleaned
func (a *App) AnalyzeCleanable() ([]models.CleanableItem, error) {
	logrus.Info("analyze cleanable requested")
	if a.reg == nil {
		logrus.Error("analyze cleanable failed: registry not initialized")
		return nil, fmt.Errorf("registry not initialized")
	}
	envs, err := a.reg.ListEnvs()
	if err != nil {
		logrus.WithError(err).Error("analyze cleanable failed to list environments")
		return nil, err
	}

	var items []models.CleanableItem
	for _, env := range envs {
		paths, _ := a.reg.ListPaths(env.ID)
		for _, p := range paths {
			if p.Type == models.PathCache && p.SizeBytes > 0 {
				items = append(items, models.CleanableItem{
					Name:        fmt.Sprintf("%s cache", env.Name),
					Path:        p.Path,
					Description: fmt.Sprintf("%s cache directory", env.Name),
					SizeBytes:   p.SizeBytes,
					Selected:    true,
					EnvKey:      env.Key,
					Category:    "cache",
					RiskLevel:   "low",
				})
			}
		}
	}

	home, _ := os.UserHomeDir()
	cacheDirs := []struct {
		path     string
		name     string
		envKey   string
		category string
		risk     string
	}{
		{filepath.Join(home, "AppData", "Local", "npm-cache"), "npm cache", "nodejs", "cache", "low"},
		{filepath.Join(home, "AppData", "Local", "pnpm-cache"), "pnpm cache", "pnpm", "cache", "low"},
		{filepath.Join(home, "AppData", "Local", "Yarn", "Cache"), "yarn cache", "yarn", "cache", "low"},
		{filepath.Join(home, ".cache", "yarn"), "yarn cache (unix)", "yarn", "cache", "low"},
		{filepath.Join(home, ".pnpm-store"), "pnpm store", "pnpm", "cache", "low"},
		{filepath.Join(home, "AppData", "Local", "pip", "Cache"), "pip cache", "python", "cache", "low"},
		{filepath.Join(home, ".cache", "pip"), "pip cache (unix)", "python", "cache", "low"},
		{filepath.Join(home, ".gradle", "caches"), "gradle caches", "java", "cache", "low"},
		{filepath.Join(home, ".m2", "repository"), "maven repository", "java", "cache", "low"},
		{filepath.Join(home, ".cargo", "registry"), "cargo registry", "rust", "cache", "low"},
		{filepath.Join(home, "go", "pkg", "mod"), "go module cache", "go", "cache", "low"},
		{filepath.Join(home, ".bun", "install", "cache"), "bun install cache", "bun", "cache", "low"},
	}

	for _, cd := range cacheDirs {
		if scanner.PathExists(cd.path) {
			size := scanner.DirSize(cd.path)
			if size > 0 {
				items = append(items, models.CleanableItem{
					Name:        cd.name,
					Path:        cd.path,
					Description: fmt.Sprintf("%s directory", cd.name),
					SizeBytes:   size,
					Selected:    true,
					EnvKey:      cd.envKey,
					Category:    cd.category,
					RiskLevel:   cd.risk,
				})
			}
		}
	}

	cwd, _ := os.Getwd()
	projectDirs := []struct {
		name     string
		subpath  string
		envKey   string
		category string
		risk     string
	}{
		{"node_modules", "node_modules", "nodejs", "deps", "medium"},
		{"Maven target", "target", "java", "build", "medium"},
	}

	for _, pd := range projectDirs {
		p := filepath.Join(cwd, pd.subpath)
		if scanner.IsDir(p) {
			size := scanner.DirSize(p)
			if size > 0 {
				items = append(items, models.CleanableItem{
					Name:        pd.name,
					Path:        p,
					Description: fmt.Sprintf("Project %s directory", pd.category),
					SizeBytes:   size,
					Selected:    false,
					EnvKey:      pd.envKey,
					Category:    pd.category,
					RiskLevel:   pd.risk,
				})
			}
		}
	}

	logrus.WithField("item_count", len(items)).Info("analyze cleanable completed")
	return items, nil
}

// CleanItems removes selected cache directories
func (a *App) CleanItems(items []models.CleanableItem) (int64, error) {
	logrus.WithField("item_count", len(items)).Info("clean items requested")
	var totalFreed int64
	allowedItems, err := a.AnalyzeCleanable()
	if err != nil {
		logrus.WithError(err).Error("clean items failed to analyze allowed items")
		return 0, err
	}
	allowed := make(map[string]models.CleanableItem, len(allowedItems))
	for _, item := range allowedItems {
		path, ok := normalizedCleanPath(item.Path)
		if ok && isSafeCleanPath(path) {
			allowed[path] = item
		}
	}
	var failures []string
	for _, item := range items {
		if !item.Selected {
			continue
		}
		path, ok := normalizedCleanPath(item.Path)
		if !ok || !isSafeCleanPath(path) {
			logrus.WithField("path", item.Path).Warn("unsafe clean path skipped")
			failures = append(failures, fmt.Sprintf("unsafe path skipped: %s", item.Path))
			continue
		}
		allowedItem, ok := allowed[path]
		if !ok {
			logrus.WithField("path", item.Path).Warn("unrecognized clean path skipped")
			failures = append(failures, fmt.Sprintf("unrecognized cleanable path skipped: %s", item.Path))
			continue
		}
		size := scanner.DirSize(path)
		if err := os.RemoveAll(path); err != nil {
			logrus.WithError(err).WithField("path", path).Error("failed to clean path")
			failures = append(failures, fmt.Sprintf("%s: %v", path, err))
			continue
		}
		logrus.WithFields(logrus.Fields{"path": path, "bytes": size, "env_key": allowedItem.EnvKey}).Info("cleaned path")
		totalFreed += size
		_ = a.reg.SaveHistory(&models.HistoryEntry{
			Action:      "clean",
			TargetEnv:   allowedItem.EnvKey,
			DetailsJSON: fmt.Sprintf(`{"path":"%s","bytes":%d}`, path, size),
			Success:     true,
			CreatedAt:   time.Now(),
		})
	}
	if len(failures) > 0 {
		logrus.WithFields(logrus.Fields{"bytes_freed": totalFreed, "failures": len(failures)}).Warn("clean items completed with failures")
		a.saveCleanMetricSnapshot(totalFreed)
		return totalFreed, fmt.Errorf("%s", strings.Join(failures, "; "))
	}
	logrus.WithField("bytes_freed", totalFreed).Info("clean items completed")
	a.saveCleanMetricSnapshot(totalFreed)
	return totalFreed, nil
}

func (a *App) saveCleanMetricSnapshot(totalFreed int64) {
	if totalFreed <= 0 || a.reg == nil {
		return
	}
	if err := a.reg.SaveMetricSnapshot(&models.MetricSnapshot{
		MetricKey:  "clean_freed_bytes",
		TargetKey:  "all",
		ValueBytes: totalFreed,
		CapturedAt: time.Now(),
	}); err != nil {
		logrus.WithError(err).WithField("metric_key", "clean_freed_bytes").Warn("failed to save metric snapshot")
	}
	if err := a.reg.PruneMetricSnapshots("clean_freed_bytes", 1000); err != nil {
		logrus.WithError(err).WithField("metric_key", "clean_freed_bytes").Warn("failed to prune metric snapshots")
	}
}

func normalizedCleanPath(path string) (string, bool) {
	if strings.TrimSpace(path) == "" {
		return "", false
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", false
	}
	return filepath.Clean(abs), true
}

func isSafeCleanPath(path string) bool {
	if !scanner.IsDir(path) {
		return false
	}
	clean := filepath.Clean(path)
	if filepath.Dir(clean) == clean {
		return false
	}
	volume := filepath.VolumeName(clean)
	if volume != "" && clean == volume+string(os.PathSeparator) {
		return false
	}
	home, _ := os.UserHomeDir()
	for _, base := range []string{home, os.TempDir()} {
		if base != "" && filepath.Clean(base) == clean {
			return false
		}
	}
	return true
}
