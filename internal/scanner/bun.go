package scanner

import (
	"devman/internal/models"
	"os"
	"path/filepath"
	"runtime"
)

type BunScanner struct{}

func (s *BunScanner) Name() string { return "Bun" }

func (s *BunScanner) Detect() ([]models.EnvInstance, []models.EnvPath, error) {
	var instances []models.EnvInstance
	var paths []models.EnvPath

	exe := FindExecutableInPath("bun")
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
		bunCache := filepath.Join(home, ".bun", "install", "cache")
		if PathExists(bunCache) {
			paths = append(paths, models.EnvPath{Type: models.PathCache, Path: bunCache, IsMovable: true})
		}
	} else {
		bunCache := filepath.Join(home, ".bun", "install", "cache")
		if PathExists(bunCache) {
			paths = append(paths, models.EnvPath{Type: models.PathCache, Path: bunCache, IsMovable: true})
		}
	}

	return instances, paths, nil
}
