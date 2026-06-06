package versionmanager

import (
	"reflect"
	"testing"
)

func TestSupportedToolsContainRequiredMetadata(t *testing.T) {
	tools := SupportedTools()

	expected := []ToolDefinition{
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

	if len(tools) != len(expected) {
		t.Fatalf("expected %d supported tools, got %d", len(expected), len(tools))
	}

	seen := make(map[string]bool, len(tools))
	for _, tool := range tools {
		if seen[tool.Key] {
			t.Fatalf("duplicate tool key %q", tool.Key)
		}
		seen[tool.Key] = true
	}

	for _, want := range expected {
		t.Run(want.Key, func(t *testing.T) {
			got, ok := ToolByKey(want.Key)
			if !ok {
				t.Fatalf("missing tool %q", want.Key)
			}
			if got.Key != want.Key {
				t.Fatalf("%s Key = %q, want %q", want.Key, got.Key, want.Key)
			}
			if got.Name != want.Name {
				t.Fatalf("%s Name = %q, want %q", want.Key, got.Name, want.Name)
			}
			if got.EnvVar != want.EnvVar {
				t.Fatalf("%s EnvVar = %q, want %q", want.Key, got.EnvVar, want.EnvVar)
			}
			if got.PrimaryExe != want.PrimaryExe {
				t.Fatalf("%s PrimaryExe = %q, want %q", want.Key, got.PrimaryExe, want.PrimaryExe)
			}
			if !reflect.DeepEqual(got.Shims, want.Shims) {
				t.Fatalf("%s Shims = %#v, want %#v", want.Key, got.Shims, want.Shims)
			}
			if !reflect.DeepEqual(got.VersionArgs, want.VersionArgs) {
				t.Fatalf("%s VersionArgs = %#v, want %#v", want.Key, got.VersionArgs, want.VersionArgs)
			}
			if got.BinSubdir != want.BinSubdir {
				t.Fatalf("%s BinSubdir = %q, want %q", want.Key, got.BinSubdir, want.BinSubdir)
			}
		})
	}
}
