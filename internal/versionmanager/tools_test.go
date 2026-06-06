package versionmanager

import "testing"

func TestSupportedToolsContainRequiredMetadata(t *testing.T) {
	tools := SupportedTools()
	if len(tools) != 4 {
		t.Fatalf("expected 4 supported tools, got %d", len(tools))
	}

	assertTool := func(key string, envVar string, primaryShim string, versionArgs []string) {
		t.Helper()
		tool, ok := ToolByKey(key)
		if !ok {
			t.Fatalf("missing tool %q", key)
		}
		if tool.EnvVar != envVar {
			t.Fatalf("%s EnvVar = %q, want %q", key, tool.EnvVar, envVar)
		}
		if len(tool.Shims) == 0 || tool.Shims[0] != primaryShim {
			t.Fatalf("%s primary shim = %#v, want %q", key, tool.Shims, primaryShim)
		}
		if len(tool.VersionArgs) != len(versionArgs) {
			t.Fatalf("%s VersionArgs = %#v, want %#v", key, tool.VersionArgs, versionArgs)
		}
		for i := range versionArgs {
			if tool.VersionArgs[i] != versionArgs[i] {
				t.Fatalf("%s VersionArgs = %#v, want %#v", key, tool.VersionArgs, versionArgs)
			}
		}
	}

	assertTool("go", "GOROOT", "go.cmd", []string{"version"})
	assertTool("node", "NODE_HOME", "node.cmd", []string{"--version"})
	assertTool("bun", "BUN_INSTALL", "bun.cmd", []string{"--version"})
	assertTool("flutter", "FLUTTER_ROOT", "flutter.cmd", []string{"--version"})
}
