package scanner

import (
	"devman/internal/models"
	"strings"
)

type ScannerDescriptor struct {
	Env            models.Env
	CustomExeNames []string
	SyncedToolKey  string
}

type scannerDescriptorProvider interface {
	Descriptor() ScannerDescriptor
}

func descriptorForScanner(s Scanner) ScannerDescriptor {
	if provider, ok := s.(scannerDescriptorProvider); ok {
		desc := provider.Descriptor()
		if desc.Env.Key == "" {
			desc.Env.Key = scannerKey(s.Name())
		}
		if desc.Env.Name == "" {
			desc.Env.Name = s.Name()
		}
		if desc.Env.Category == "" {
			desc.Env.Category = models.CategoryTool
		}
		if desc.SyncedToolKey == "" {
			desc.SyncedToolKey = syncedToolKeyForEnvKey(desc.Env.Key)
		}
		return desc
	}
	desc := ScannerDescriptor{
		Env: models.Env{
			Name:     s.Name(),
			Key:      scannerKey(s.Name()),
			Category: models.CategoryTool,
		},
	}
	desc.SyncedToolKey = syncedToolKeyForEnvKey(desc.Env.Key)
	return desc
}

func scannerKey(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func syncedToolKeyForEnvKey(envKey string) string {
	switch strings.TrimSpace(envKey) {
	case "go", "bun", "flutter":
		return envKey
	case "nodejs":
		return "node"
	default:
		return ""
	}
}

func (s *NodeScanner) Descriptor() ScannerDescriptor {
	return ScannerDescriptor{
		Env:            models.Env{Name: "Node.js", Key: "nodejs", Category: models.CategoryRuntime, Icon: "⚡", Description: "JavaScript runtime", Website: "https://nodejs.org"},
		CustomExeNames: []string{"node", "node.exe"},
		SyncedToolKey:  "node",
	}
}

func (s *PythonScanner) Descriptor() ScannerDescriptor {
	return ScannerDescriptor{
		Env:            models.Env{Name: "Python", Key: "python", Category: models.CategoryRuntime, Icon: "🐍", Description: "Python interpreter", Website: "https://python.org"},
		CustomExeNames: []string{"python", "python3", "python.exe", "py"},
	}
}

func (s *JavaScanner) Descriptor() ScannerDescriptor {
	return ScannerDescriptor{
		Env:            models.Env{Name: "Java", Key: "java", Category: models.CategoryRuntime, Icon: "☕", Description: "Java Development Kit", Website: "https://openjdk.org"},
		CustomExeNames: []string{"java", "java.exe"},
	}
}

func (s *GoScanner) Descriptor() ScannerDescriptor {
	return ScannerDescriptor{
		Env:            models.Env{Name: "Go", Key: "go", Category: models.CategoryRuntime, Icon: "🦀", Description: "Go programming language", Website: "https://go.dev"},
		CustomExeNames: []string{"go", "go.exe"},
		SyncedToolKey:  "go",
	}
}

func (s *FlutterScanner) Descriptor() ScannerDescriptor {
	return ScannerDescriptor{
		Env:           models.Env{Name: "Flutter", Key: "flutter", Category: models.CategorySDK, Icon: "🎯", Description: "Flutter SDK", Website: "https://flutter.dev"},
		SyncedToolKey: "flutter",
	}
}

func (s *RustScanner) Descriptor() ScannerDescriptor {
	return ScannerDescriptor{
		Env: models.Env{Name: "Rust", Key: "rust", Category: models.CategoryRuntime, Icon: "🦾", Description: "Rust toolchain", Website: "https://rust-lang.org"},
	}
}

func (s *DockerScanner) Descriptor() ScannerDescriptor {
	return ScannerDescriptor{
		Env: models.Env{Name: "Docker", Key: "docker", Category: models.CategoryTool, Icon: "🐳", Description: "Container runtime", Website: "https://docker.com"},
	}
}

func (s *PnpmScanner) Descriptor() ScannerDescriptor {
	return ScannerDescriptor{
		Env: models.Env{Name: "pnpm", Key: "pnpm", Category: models.CategoryTool, Icon: "📦", Description: "Fast, disk space efficient package manager", Website: "https://pnpm.io"},
	}
}

func (s *YarnScanner) Descriptor() ScannerDescriptor {
	return ScannerDescriptor{
		Env: models.Env{Name: "Yarn", Key: "yarn", Category: models.CategoryTool, Icon: "🧶", Description: "Yarn package manager", Website: "https://yarnpkg.com"},
	}
}

func (s *BunScanner) Descriptor() ScannerDescriptor {
	return ScannerDescriptor{
		Env:           models.Env{Name: "Bun", Key: "bun", Category: models.CategoryRuntime, Icon: "🥟", Description: "Bun JavaScript runtime", Website: "https://bun.sh"},
		SyncedToolKey: "bun",
	}
}
