package versionmanager

type ToolDefinition struct {
	Key         string
	Name        string
	EnvVar      string
	PrimaryExe  string
	Shims       []string
	VersionArgs []string
	BinSubdir   string
}

func SupportedTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Key:         "go",
			Name:        "Go",
			EnvVar:      "GOROOT",
			PrimaryExe:  "go.exe",
			Shims:       []string{"go.cmd"},
			VersionArgs: []string{"version"},
			BinSubdir:   "bin",
		},
		{
			Key:         "node",
			Name:        "Node.js",
			EnvVar:      "NODE_HOME",
			PrimaryExe:  "node.exe",
			Shims:       []string{"node.cmd", "npm.cmd", "npx.cmd"},
			VersionArgs: []string{"--version"},
			BinSubdir:   "",
		},
		{
			Key:         "bun",
			Name:        "Bun",
			EnvVar:      "BUN_INSTALL",
			PrimaryExe:  "bun.exe",
			Shims:       []string{"bun.cmd"},
			VersionArgs: []string{"--version"},
			BinSubdir:   "",
		},
		{
			Key:         "flutter",
			Name:        "Flutter",
			EnvVar:      "FLUTTER_ROOT",
			PrimaryExe:  "flutter.bat",
			Shims:       []string{"flutter.cmd", "dart.cmd"},
			VersionArgs: []string{"--version"},
			BinSubdir:   "bin",
		},
	}
}

func ToolByKey(key string) (ToolDefinition, bool) {
	for _, tool := range SupportedTools() {
		if tool.Key == key {
			return tool, true
		}
	}
	return ToolDefinition{}, false
}
