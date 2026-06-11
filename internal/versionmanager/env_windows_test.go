//go:build windows

package versionmanager

import "testing"

func TestPrioritizePathEntryMovesExistingEntryToFront(t *testing.T) {
	current := `C:\Tools\runtime;%DEVMAN_HOME%\shims;C:\Windows\System32`

	got, changed := prioritizePathEntry(current, `%DEVMAN_HOME%\shims`)

	if !changed {
		t.Fatal("expected PATH to change when the shim entry exists but is not first")
	}
	want := `%DEVMAN_HOME%\shims;C:\Tools\runtime;C:\Windows\System32`
	if got != want {
		t.Fatalf("prioritized PATH = %q, want %q", got, want)
	}
}

func TestPrioritizePathEntryLeavesFirstEntryUnchanged(t *testing.T) {
	current := `%DEVMAN_HOME%\shims;C:\Tools\runtime`

	got, changed := prioritizePathEntry(current, `%DEVMAN_HOME%\shims`)

	if changed {
		t.Fatal("expected PATH to remain unchanged when the shim entry is already first")
	}
	if got != current {
		t.Fatalf("PATH = %q, want %q", got, current)
	}
}
