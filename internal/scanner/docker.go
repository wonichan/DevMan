package scanner

import (
	"devman/internal/models"
	"os"
	"path/filepath"
	"runtime"
)

type DockerScanner struct{}

func (s *DockerScanner) Name() string { return "Docker" }

func (s *DockerScanner) Detect() ([]models.EnvInstance, []models.EnvPath, error) {
	var instances []models.EnvInstance
	var paths []models.EnvPath

	exe := FindExecutableInPath("docker")
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
	dockerCache := filepath.Join(home, ".docker")
	if PathExists(dockerCache) {
		paths = append(paths, models.EnvPath{Type: models.PathCache, Path: dockerCache, IsMovable: true})
	}

	if runtime.GOOS == "windows" {
		desktopData := filepath.Join(home, "AppData", "Local", "Docker")
		if PathExists(desktopData) {
			paths = append(paths, models.EnvPath{Type: models.PathData, Path: desktopData, IsMovable: true})
		}
	}

	return instances, paths, nil
}
