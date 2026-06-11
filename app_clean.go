package main

import (
	"devman/internal/models"
	"devman/internal/scanner"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// AnalyzeCleanable returns items that can be safely cleaned.
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

// CleanItems removes selected cache directories.
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
