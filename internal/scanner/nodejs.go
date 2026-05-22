package scanner

import (
	"devman/internal/models"
	"os"
	"path/filepath"
	"strings"
)

type NodeScanner struct{}

func (s *NodeScanner) Name() string { return "Node.js" }

func (s *NodeScanner) Detect() ([]models.EnvInstance, []models.EnvPath, error) {
	var instances []models.EnvInstance
	var paths []models.EnvPath

	// Detect via executable
	exe := FindExecutableInPath("node")
	if exe == "" {
		// Fallback common paths
		for _, base := range CommonWindowsPaths() {
			candidates := []string{
				filepath.Join(base, "nodejs", "node.exe"),
				filepath.Join(base, "node", "node.exe"),
			}
			for _, c := range candidates {
				if PathExists(c) {
					exe = c
					break
				}
			}
			if exe != "" {
				break
			}
		}
	}

	if exe != "" {
		installDir := filepath.Dir(exe)
		version := s.readVersion(exe)
		instances = append(instances, models.EnvInstance{
			Version:     version,
			InstallPath: installDir,
			IsDefault:   true,
			IsActive:    true,
			Source:      "system",
		})
		paths = append(paths, models.EnvPath{
			Type:      models.PathInstall,
			Path:      installDir,
			IsMovable: true,
		})

		// Cache paths
		home, _ := os.UserHomeDir()
		npmCache := filepath.Join(home, "AppData", "Local", "npm-cache")
		if PathExists(npmCache) {
			paths = append(paths, models.EnvPath{Type: models.PathCache, Path: npmCache, IsMovable: true})
		}
		npmGlobal := filepath.Join(home, "AppData", "Roaming", "npm")
		if PathExists(npmGlobal) {
			paths = append(paths, models.EnvPath{Type: models.PathDeps, Path: npmGlobal, IsMovable: true})
		}
	}

	return instances, paths, nil
}

func (s *NodeScanner) readVersion(exe string) string {
	// In real implementation, run `node --version` and parse output
	// For now return placeholder
	return "detected"
}

func (s *NodeScanner) detectYarnCache() string {
	home, _ := os.UserHomeDir()
	c := filepath.Join(home, "AppData", "Local", "Yarn", "Cache")
	if PathExists(c) {
		return c
	}
	return ""
}

func (s *NodeScanner) detectPnpmCache() string {
	home, _ := os.UserHomeDir()
	c := filepath.Join(home, "AppData", "Local", "pnpm-store")
	if PathExists(c) {
		return c
	}
	c = filepath.Join(home, ".pnpm-store")
	if PathExists(c) {
		return c
	}
	return ""
}

type PythonScanner struct{}

func (s *PythonScanner) Name() string { return "Python" }

func (s *PythonScanner) Detect() ([]models.EnvInstance, []models.EnvPath, error) {
	var instances []models.EnvInstance
	var paths []models.EnvPath

	names := []string{"python", "python3", "py"}
	for _, name := range names {
		exe := FindExecutableInPath(name)
		if exe != "" {
			installDir := filepath.Dir(exe)
			// On Windows, Python is usually in C:\Python3xx or exe is in Scripts subdir
			if strings.Contains(strings.ToLower(installDir), "scripts") {
				installDir = filepath.Dir(installDir)
			}
			instances = append(instances, models.EnvInstance{
				Version:     "detected",
				InstallPath: installDir,
				IsDefault:   name == "python",
				IsActive:    true,
				Source:      "system",
			})
			paths = append(paths, models.EnvPath{Type: models.PathInstall, Path: installDir, IsMovable: true})
		}
	}

	if len(instances) == 0 {
		// Fallback common paths
		for _, base := range CommonWindowsPaths() {
			candidates := []string{
				filepath.Join(base, "Python311"),
				filepath.Join(base, "Python310"),
				filepath.Join(base, "Python312"),
				filepath.Join(base, "Python39"),
			}
			for _, c := range candidates {
				if IsDir(c) {
					instances = append(instances, models.EnvInstance{
						Version:     filepath.Base(c),
						InstallPath: c,
						IsDefault:   false,
						IsActive:    false,
						Source:      "system",
					})
					paths = append(paths, models.EnvPath{Type: models.PathInstall, Path: c, IsMovable: true})
				}
			}
		}
	}

	// Pip cache
	home, _ := os.UserHomeDir()
	pipCache := filepath.Join(home, "AppData", "Local", "pip", "Cache")
	if PathExists(pipCache) {
		paths = append(paths, models.EnvPath{Type: models.PathCache, Path: pipCache, IsMovable: true})
	}

	return instances, paths, nil
}

type JavaScanner struct{}

func (s *JavaScanner) Name() string { return "Java" }

func (s *JavaScanner) Detect() ([]models.EnvInstance, []models.EnvPath, error) {
	var instances []models.EnvInstance
	var paths []models.EnvPath

	exe := FindExecutableInPath("java")
	if exe != "" {
		installDir := filepath.Dir(exe)
		if strings.Contains(strings.ToLower(installDir), "bin") {
			installDir = filepath.Dir(installDir)
		}
		instances = append(instances, models.EnvInstance{
			Version:     "detected",
			InstallPath: installDir,
			IsDefault:   true,
			IsActive:    true,
			Source:      "system",
		})
		paths = append(paths, models.EnvPath{Type: models.PathInstall, Path: installDir, IsMovable: true})
	}

	// Fallback paths
	for _, base := range CommonWindowsPaths() {
		candidates := []string{
			filepath.Join(base, "Java", "jdk-17"),
			filepath.Join(base, "Java", "jdk-11"),
			filepath.Join(base, "Java", "jdk-21"),
		}
		for _, c := range candidates {
			if IsDir(c) {
				instances = append(instances, models.EnvInstance{
					Version:     filepath.Base(c),
					InstallPath: c,
					IsDefault:   false,
					IsActive:    false,
					Source:      "system",
				})
				paths = append(paths, models.EnvPath{Type: models.PathInstall, Path: c, IsMovable: true})
			}
		}
	}

	// Gradle cache
	home, _ := os.UserHomeDir()
	gradleCache := filepath.Join(home, ".gradle")
	if PathExists(gradleCache) {
		paths = append(paths, models.EnvPath{Type: models.PathCache, Path: gradleCache, IsMovable: true})
	}
	// Maven cache
	m2 := filepath.Join(home, ".m2")
	if PathExists(m2) {
		paths = append(paths, models.EnvPath{Type: models.PathCache, Path: m2, IsMovable: true})
	}

	return instances, paths, nil
}

type GoScanner struct{}

func (s *GoScanner) Name() string { return "Go" }

func (s *GoScanner) Detect() ([]models.EnvInstance, []models.EnvPath, error) {
	var instances []models.EnvInstance
	var paths []models.EnvPath

	exe := FindExecutableInPath("go")
	if exe != "" {
		installDir := filepath.Dir(exe)
		if strings.Contains(strings.ToLower(installDir), "bin") {
			installDir = filepath.Dir(installDir)
		}
		instances = append(instances, models.EnvInstance{
			Version:     "detected",
			InstallPath: installDir,
			IsDefault:   true,
			IsActive:    true,
			Source:      "system",
		})
		paths = append(paths, models.EnvPath{Type: models.PathInstall, Path: installDir, IsMovable: true})
	}

	// Go module cache
	home, _ := os.UserHomeDir()
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		goPath = filepath.Join(home, "go")
	}
	modCache := filepath.Join(goPath, "pkg", "mod")
	if PathExists(modCache) {
		paths = append(paths, models.EnvPath{Type: models.PathCache, Path: modCache, IsMovable: true})
	}

	return instances, paths, nil
}

type FlutterScanner struct{}

func (s *FlutterScanner) Name() string { return "Flutter" }

func (s *FlutterScanner) Detect() ([]models.EnvInstance, []models.EnvPath, error) {
	var instances []models.EnvInstance
	var paths []models.EnvPath

	exe := FindExecutableInPath("flutter")
	if exe != "" {
		installDir := filepath.Dir(exe)
		if strings.Contains(strings.ToLower(installDir), "bin") {
			installDir = filepath.Dir(installDir)
		}
		instances = append(instances, models.EnvInstance{
			Version:     "detected",
			InstallPath: installDir,
			IsDefault:   true,
			IsActive:    true,
			Source:      "system",
		})
		paths = append(paths, models.EnvPath{Type: models.PathInstall, Path: installDir, IsMovable: true})
	} else {
		// Fallback
		for _, base := range CommonWindowsPaths() {
			c := filepath.Join(base, "flutter")
			if IsDir(c) {
				instances = append(instances, models.EnvInstance{
					Version:     "detected",
					InstallPath: c,
					IsDefault:   true,
					IsActive:    false,
					Source:      "system",
				})
				paths = append(paths, models.EnvPath{Type: models.PathInstall, Path: c, IsMovable: true})
				break
			}
		}
	}

	// Flutter build cache
	if len(instances) > 0 {
		cache := filepath.Join(instances[0].InstallPath, "bin", "cache")
		if PathExists(cache) {
			paths = append(paths, models.EnvPath{Type: models.PathCache, Path: cache, IsMovable: true})
		}
	}

	return instances, paths, nil
}

type RustScanner struct{}

func (s *RustScanner) Name() string { return "Rust" }

func (s *RustScanner) Detect() ([]models.EnvInstance, []models.EnvPath, error) {
	var instances []models.EnvInstance
	var paths []models.EnvPath

	exe := FindExecutableInPath("rustc")
	if exe != "" {
		installDir := filepath.Dir(exe)
		if strings.Contains(strings.ToLower(installDir), "bin") {
			installDir = filepath.Dir(installDir)
		}
		instances = append(instances, models.EnvInstance{
			Version:     "detected",
			InstallPath: installDir,
			IsDefault:   true,
			IsActive:    true,
			Source:      "system",
		})
		paths = append(paths, models.EnvPath{Type: models.PathInstall, Path: installDir, IsMovable: true})
	}

	// Cargo cache
	home, _ := os.UserHomeDir()
	cargoCache := filepath.Join(home, ".cargo")
	if PathExists(cargoCache) {
		paths = append(paths, models.EnvPath{Type: models.PathCache, Path: cargoCache, IsMovable: true})
	}
	cargoRegistry := filepath.Join(home, ".cargo", "registry")
	if PathExists(cargoRegistry) {
		paths = append(paths, models.EnvPath{Type: models.PathDeps, Path: cargoRegistry, IsMovable: true})
	}

	return instances, paths, nil
}
