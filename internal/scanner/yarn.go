package scanner

import (
	"devman/internal/models"
	"os"
	"path/filepath"
	"runtime"
)

type YarnScanner struct{}

func (s *YarnScanner) Name() string { return "Yarn" }

func (s *YarnScanner) Detect() ([]models.EnvInstance, []models.EnvPath, error) {
	var instances []models.EnvInstance
	var paths []models.EnvPath

	exe := FindExecutableInPath("yarn")
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
		yarnCache := filepath.Join(home, "AppData", "Local", "Yarn", "Cache")
		if PathExists(yarnCache) {
			paths = append(paths, models.EnvPath{Type: models.PathCache, Path: yarnCache, IsMovable: true})
		}
	} else {
		yarnCache := filepath.Join(home, ".cache", "yarn")
		if PathExists(yarnCache) {
			paths = append(paths, models.EnvPath{Type: models.PathCache, Path: yarnCache, IsMovable: true})
		}
	}

	return instances, paths, nil
}
