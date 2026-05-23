package scanner

import (
	"devman/internal/models"
	"devman/internal/registry"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type Scanner interface {
	Name() string
	Detect() ([]models.EnvInstance, []models.EnvPath, error)
}

type ScanOptions struct {
	CustomScanPaths []string
}

type Engine struct {
	reg      *registry.Registry
	scanners []Scanner
}

func NewEngine(reg *registry.Registry) *Engine {
	return &Engine{
		reg: reg,
		scanners: []Scanner{
			&NodeScanner{},
			&PythonScanner{},
			&JavaScanner{},
			&GoScanner{},
			&FlutterScanner{},
			&RustScanner{},
			&DockerScanner{},
			&PnpmScanner{},
			&YarnScanner{},
			&BunScanner{},
		},
	}
}

func (e *Engine) Register(s Scanner) {
	e.scanners = append(e.scanners, s)
}

func (e *Engine) ScanAll() ([]models.EnvSummary, error) {
	return e.ScanAllWithOptions(ScanOptions{})
}

func (e *Engine) ScanAllWithOptions(opts ScanOptions) ([]models.EnvSummary, error) {
	start := time.Now()
	logrus.WithFields(logrus.Fields{"scanner_count": len(e.scanners), "custom_scan_paths": len(opts.CustomScanPaths)}).Info("environment scan started")
	var summaries []models.EnvSummary
	for _, s := range e.scanners {
		scannerStart := time.Now()
		entry := logrus.WithField("scanner", s.Name())
		entry.Info("scanner started")
		instances, paths, err := s.Detect()
		if err != nil {
			entry.WithError(err).Error("scanner failed")
			continue
		}

		if len(opts.CustomScanPaths) > 0 {
			customInst, customPaths := detectCustomPaths(s, opts.CustomScanPaths)
			instances = append(instances, customInst...)
			paths = append(paths, customPaths...)
		}

		if len(instances) == 0 {
			entry.WithField("duration_ms", time.Since(scannerStart).Milliseconds()).Info("scanner completed with no instances")
			continue
		}

		// Save env metadata
		env := modelsForScanner(s)
		if err := e.reg.SaveEnv(&env); err != nil {
			entry.WithError(err).WithField("env_key", env.Key).Error("failed to save scanned environment")
			continue
		}

		// Clear old data
		if err := e.reg.ClearInstances(env.ID); err != nil {
			entry.WithError(err).WithField("env_id", env.ID).Warn("failed to clear old instances")
		}
		if err := e.reg.ClearPaths(env.ID); err != nil {
			entry.WithError(err).WithField("env_id", env.ID).Warn("failed to clear old paths")
		}

		// Save instances
		for i := range instances {
			instances[i].EnvID = env.ID
			if err := e.reg.SaveInstance(&instances[i]); err != nil {
				entry.WithError(err).WithField("install_path", instances[i].InstallPath).Warn("failed to save scanned instance")
			}
		}

		// Save paths and compute sizes
		for i := range paths {
			paths[i].EnvID = env.ID
			if paths[i].SizeBytes == 0 {
				paths[i].SizeBytes = DirSize(paths[i].Path)
			}
			if err := e.reg.SavePath(&paths[i]); err != nil {
				entry.WithError(err).WithField("path", paths[i].Path).Warn("failed to save scanned path")
			}
		}

		totalSize := int64(0)
		for _, p := range paths {
			totalSize += p.SizeBytes
		}

		health := models.HealthHealthy
		if totalSize > 5*1024*1024*1024 {
			health = models.HealthWarning
		}

		summaries = append(summaries, models.EnvSummary{
			Env:       env,
			Instances: instances,
			Paths:     paths,
			TotalSize: totalSize,
			Health:    health,
		})
		entry.WithFields(logrus.Fields{"instances": len(instances), "paths": len(paths), "total_size": totalSize, "duration_ms": time.Since(scannerStart).Milliseconds()}).Info("scanner completed")
	}
	logrus.WithFields(logrus.Fields{"summary_count": len(summaries), "duration_ms": time.Since(start).Milliseconds()}).Info("environment scan completed")
	return summaries, nil
}

func modelsForScanner(s Scanner) models.Env {
	switch s.(type) {
	case *NodeScanner:
		return models.Env{Name: "Node.js", Key: "nodejs", Category: models.CategoryRuntime, Icon: "⚡", Description: "JavaScript runtime", Website: "https://nodejs.org"}
	case *PythonScanner:
		return models.Env{Name: "Python", Key: "python", Category: models.CategoryRuntime, Icon: "🐍", Description: "Python interpreter", Website: "https://python.org"}
	case *JavaScanner:
		return models.Env{Name: "Java", Key: "java", Category: models.CategoryRuntime, Icon: "☕", Description: "Java Development Kit", Website: "https://openjdk.org"}
	case *GoScanner:
		return models.Env{Name: "Go", Key: "go", Category: models.CategoryRuntime, Icon: "🦀", Description: "Go programming language", Website: "https://go.dev"}
	case *FlutterScanner:
		return models.Env{Name: "Flutter", Key: "flutter", Category: models.CategorySDK, Icon: "🎯", Description: "Flutter SDK", Website: "https://flutter.dev"}
	case *RustScanner:
		return models.Env{Name: "Rust", Key: "rust", Category: models.CategoryRuntime, Icon: "🦾", Description: "Rust toolchain", Website: "https://rust-lang.org"}
	case *DockerScanner:
		return models.Env{Name: "Docker", Key: "docker", Category: models.CategoryTool, Icon: "🐳", Description: "Container runtime", Website: "https://docker.com"}
	case *PnpmScanner:
		return models.Env{Name: "pnpm", Key: "pnpm", Category: models.CategoryTool, Icon: "📦", Description: "Fast, disk space efficient package manager", Website: "https://pnpm.io"}
	case *YarnScanner:
		return models.Env{Name: "Yarn", Key: "yarn", Category: models.CategoryTool, Icon: "🧶", Description: "Yarn package manager", Website: "https://yarnpkg.com"}
	case *BunScanner:
		return models.Env{Name: "Bun", Key: "bun", Category: models.CategoryRuntime, Icon: "🥟", Description: "Bun JavaScript runtime", Website: "https://bun.sh"}
	default:
		return models.Env{Name: s.Name(), Key: s.Name(), Category: models.CategoryTool}
	}
}

func detectCustomPaths(s Scanner, customPaths []string) ([]models.EnvInstance, []models.EnvPath) {
	var instances []models.EnvInstance
	var paths []models.EnvPath

	var exeNames []string
	switch s.(type) {
	case *NodeScanner:
		exeNames = []string{"node", "node.exe"}
	case *PythonScanner:
		exeNames = []string{"python", "python3", "python.exe", "py"}
	case *JavaScanner:
		exeNames = []string{"java", "java.exe"}
	case *GoScanner:
		exeNames = []string{"go", "go.exe"}
	default:
		return instances, paths
	}

	for _, cp := range customPaths {
		cp = strings.TrimSpace(cp)
		if cp == "" {
			continue
		}
		fullPath := filepath.Join(cp, "bin")
		for _, name := range exeNames {
			candidate := filepath.Join(fullPath, name)
			if PathExists(candidate) {
				instances = append(instances, models.EnvInstance{
					Version:     "custom path",
					InstallPath: fullPath,
					IsDefault:   false,
					IsActive:    true,
					Source:      "custom",
				})
				paths = append(paths, models.EnvPath{
					Type:      models.PathInstall,
					Path:      fullPath,
					IsMovable: true,
				})
				break
			}
			candidate = filepath.Join(cp, name)
			if PathExists(candidate) {
				instances = append(instances, models.EnvInstance{
					Version:     "custom path",
					InstallPath: cp,
					IsDefault:   false,
					IsActive:    true,
					Source:      "custom",
				})
				paths = append(paths, models.EnvPath{
					Type:      models.PathInstall,
					Path:      cp,
					IsMovable: true,
				})
				break
			}
		}
	}

	return instances, paths
}
