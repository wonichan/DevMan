package main

import (
	"devman/internal/models"
	"devman/internal/scanner"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// ScanAll runs all scanners and returns summaries.
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
