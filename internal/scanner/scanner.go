package scanner

import (
	"devman/internal/models"
	"devman/internal/registry"
)

type Scanner interface {
	Name() string
	Detect() ([]models.EnvInstance, []models.EnvPath, error)
}

type Engine struct {
	reg     *registry.Registry
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
		},
	}
}

func (e *Engine) Register(s Scanner) {
	e.scanners = append(e.scanners, s)
}

func (e *Engine) ScanAll() ([]models.EnvSummary, error) {
	var summaries []models.EnvSummary
	for _, s := range e.scanners {
		instances, paths, err := s.Detect()
		if err != nil {
			continue
		}
		if len(instances) == 0 {
			continue
		}

		// Save env metadata
		env := modelsForScanner(s)
		if err := e.reg.SaveEnv(&env); err != nil {
			continue
		}

		// Clear old data
		_ = e.reg.ClearInstances(env.ID)
		_ = e.reg.ClearPaths(env.ID)

		// Save instances
		for i := range instances {
			instances[i].EnvID = env.ID
			_ = e.reg.SaveInstance(&instances[i])
		}

		// Save paths and compute sizes
		for i := range paths {
			paths[i].EnvID = env.ID
			if paths[i].SizeBytes == 0 {
				paths[i].SizeBytes = DirSize(paths[i].Path)
			}
			_ = e.reg.SavePath(&paths[i])
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
	}
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
	default:
		return models.Env{Name: s.Name(), Key: s.Name(), Category: models.CategoryTool}
	}
}
