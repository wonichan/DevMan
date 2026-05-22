package scanner

import (
	"devman/internal/models"
	"os"
	"path/filepath"
	"runtime"
)

type PnpmScanner struct{}

func (s *PnpmScanner) Name() string { return "pnpm" }

func (s *PnpmScanner) Detect() ([]models.EnvInstance, []models.EnvPath, error) {
	var instances []models.EnvInstance
	var paths []models.EnvPath

	exe := FindExecutableInPath("pnpm")
	if exe == "" {
		return instances, paths, nil
	}

	installDir := filepath.Dir(exe)
	if runtime.GOOS == "windows" {
		if filepath.Base(installDir) == "bin" || filepath.Base(installDir) == "cmd" {
			installDir = filepath.Dir(installDir)
		}
	}

	instances = append(instances, models.EnvInstance{
		Version:     ExecutableVersion(exe, "--version"),
		InstallPath: installDir,
		IsDefault:   true,
		IsActive:    true,
		Source:      "system",
	})
	paths = append(paths, models.EnvPath{Type: models.PathInstall, Path: installDir, IsMovable: true})

	home, _ := os.UserHomeDir()
	if runtime.GOOS == "windows" {
		pnpmStore := filepath.Join(home, "AppData", "Local", "pnpm-store")
		if PathExists(pnpmStore) {
			paths = append(paths, models.EnvPath{Type: models.PathCache, Path: pnpmStore, IsMovable: true})
		}
		pnpmCache := filepath.Join(home, "AppData", "Local", "pnpm-cache")
		if PathExists(pnpmCache) {
			paths = append(paths, models.EnvPath{Type: models.PathCache, Path: pnpmCache, IsMovable: true})
		}
	} else {
		pnpmStore := filepath.Join(home, ".pnpm-store")
		if PathExists(pnpmStore) {
			paths = append(paths, models.EnvPath{Type: models.PathCache, Path: pnpmStore, IsMovable: true})
		}
		pnpmCache := filepath.Join(home, ".cache", "pnpm")
		if PathExists(pnpmCache) {
			paths = append(paths, models.EnvPath{Type: models.PathCache, Path: pnpmCache, IsMovable: true})
		}
	}

	return instances, paths, nil
}
