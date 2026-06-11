package main

import (
	"context"
	"devman/internal/migrator"
	"devman/internal/registry"
	"devman/internal/scanner"
	"devman/internal/versionmanager"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
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

	// Portable mode keeps devman.db next to the executable.
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
