package main

import (
	"context"
	"devman/internal/migrator"
	"devman/internal/models"
	"devman/internal/registry"
	"devman/internal/scanner"
	"devman/internal/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx      context.Context
	reg      *registry.Registry
	engine   *scanner.Engine
	migrator *migrator.Engine
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Ensure db directory exists (portable mode: same dir as exe)
	if exe, err := os.Executable(); err == nil {
		dbDir := filepath.Dir(exe)
		os.MkdirAll(dbDir, 0755)
	}

	reg, err := registry.Open()
	if err != nil {
		fmt.Println("Failed to open registry:", err)
		return
	}
	a.reg = reg
	a.engine = scanner.NewEngine(reg)
	a.migrator = migrator.New(reg)
}

func (a *App) shutdown(ctx context.Context) {
	if a.reg != nil {
		a.reg.Close()
	}
}

// ScanAll runs all scanners and returns summaries
func (a *App) ScanAll() ([]models.EnvSummary, error) {
	if a.engine == nil {
		return nil, fmt.Errorf("engine not initialized")
	}
	settings, err := a.GetSettings()
	if err != nil {
		return nil, err
	}
	return a.engine.ScanAllWithOptions(scanner.ScanOptions{CustomScanPaths: settings.CustomScanPaths})
}

// GetSettings returns the current application settings
func (a *App) GetSettings() (*models.AppSettings, error) {
	if a.reg == nil {
		return nil, fmt.Errorf("registry not initialized")
	}
	return a.reg.GetSettings()
}

// SaveSettings persists application settings
func (a *App) SaveSettings(settings models.AppSettings) error {
	if a.reg == nil {
		return fmt.Errorf("registry not initialized")
	}
	return a.reg.SaveSettings(&settings)
}

// GetEnvs returns all stored environments
func (a *App) GetEnvs() ([]models.Env, error) {
	if a.reg == nil {
		return nil, fmt.Errorf("registry not initialized")
	}
	return a.reg.ListEnvs()
}

// GetEnvSummary returns full summary for an env
func (a *App) GetEnvSummary(key string) (*models.EnvSummary, error) {
	if a.reg == nil {
		return nil, fmt.Errorf("registry not initialized")
	}
	env, err := a.reg.GetEnvByKey(key)
	if err != nil || env == nil {
		return nil, fmt.Errorf("env not found: %s", key)
	}
	instances, _ := a.reg.ListInstances(env.ID)
	paths, _ := a.reg.ListPaths(env.ID)

	totalSize := int64(0)
	for _, p := range paths {
		totalSize += p.SizeBytes
	}

	health := models.HealthHealthy
	if totalSize > 5*1024*1024*1024 {
		health = models.HealthWarning
	}

	return &models.EnvSummary{
		Env:       *env,
		Instances: instances,
		Paths:     paths,
		TotalSize: totalSize,
		Health:    health,
	}, nil
}

// Migrate moves an env to target directory
func (a *App) Migrate(envID int64, targetDir string, useJunction bool) (*migrator.MigrationResult, error) {
	if a.migrator == nil {
		return nil, fmt.Errorf("migrator not initialized")
	}

	emitProgress := func(progress models.MigrationProgress) {
		if a.ctx != nil {
			wailsRuntime.EventsEmit(a.ctx, "migration:progress", progress)
		}
	}
	return a.migrator.MigrateWithProgress(envID, targetDir, useJunction, emitProgress)
}

// GetDiskInfo returns disk usage info (stub on Linux)
func (a *App) GetDiskInfo() ([]models.DiskInfo, error) {
	return utils.GetDiskInfo()
}

// Platform-specific disk info removed — now in utils package
func (a *App) GetHistory(limit int) ([]models.HistoryEntry, error) {
	if a.reg == nil {
		return nil, fmt.Errorf("registry not initialized")
	}
	return a.reg.GetHistory(limit)
}

// AnalyzeCleanable returns items that can be safely cleaned
func (a *App) AnalyzeCleanable() ([]models.CleanableItem, error) {
	if a.reg == nil {
		return nil, fmt.Errorf("registry not initialized")
	}
	envs, err := a.reg.ListEnvs()
	if err != nil {
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

	return items, nil
}

// CleanItems removes selected cache directories
func (a *App) CleanItems(items []models.CleanableItem) (int64, error) {
	var totalFreed int64
	allowedItems, err := a.AnalyzeCleanable()
	if err != nil {
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
			failures = append(failures, fmt.Sprintf("unsafe path skipped: %s", item.Path))
			continue
		}
		allowedItem, ok := allowed[path]
		if !ok {
			failures = append(failures, fmt.Sprintf("unrecognized cleanable path skipped: %s", item.Path))
			continue
		}
		size := scanner.DirSize(path)
		if err := os.RemoveAll(path); err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", path, err))
			continue
		}
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
		return totalFreed, fmt.Errorf("%s", strings.Join(failures, "; "))
	}
	return totalFreed, nil
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
