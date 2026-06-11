package main

import (
	"devman/internal/models"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// GetSettings returns the current application settings.
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

// SaveSettings persists application settings.
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

// GetEnvs returns all stored environments.
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

// GetEnvSummary returns full summary for an env.
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

// GetMetricSnapshots returns metric snapshots for trend cards.
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
