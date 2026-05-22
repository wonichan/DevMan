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
	"time"
)

type App struct {
	ctx     context.Context
	reg     *registry.Registry
	engine  *scanner.Engine
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
	return a.engine.ScanAll()
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
	return a.migrator.Migrate(envID, targetDir, useJunction)
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
					Description: fmt.Sprintf("%s 缓存目录", env.Name),
					SizeBytes:   p.SizeBytes,
					Selected:    true,
					EnvKey:      env.Key,
				})
			}
		}
	}
	return items, nil
}

// CleanItems removes selected cache directories
func (a *App) CleanItems(items []models.CleanableItem) (int64, error) {
	var totalFreed int64
	for _, item := range items {
		if !item.Selected {
			continue
		}
		size := scanner.DirSize(item.Path)
		if err := os.RemoveAll(item.Path); err != nil {
			continue
		}
		totalFreed += size
		_ = a.reg.SaveHistory(&models.HistoryEntry{
			Action:    "clean",
			TargetEnv: item.EnvKey,
			DetailsJSON: fmt.Sprintf(`{"path":"%s","bytes":%d}`, item.Path, size),
			Success:   true,
			CreatedAt: time.Now(),
		})
	}
	return totalFreed, nil
}

