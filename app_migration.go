package main

import (
	"devman/internal/migrator"
	"devman/internal/models"
	"fmt"

	"github.com/sirupsen/logrus"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// Migrate moves an env to target directory.
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
