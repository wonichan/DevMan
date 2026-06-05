package scanner

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDirSize(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some files
	f1 := filepath.Join(tmpDir, "file1.txt")
	os.WriteFile(f1, []byte("hello world"), 0644)

	subDir := filepath.Join(tmpDir, "subdir")
	os.MkdirAll(subDir, 0755)
	f2 := filepath.Join(subDir, "file2.txt")
	os.WriteFile(f2, []byte("test content here"), 0644)

	size := DirSize(tmpDir)
	if size == 0 {
		t.Error("dir size should not be zero")
	}
	expected := int64(11 + 17) // "hello world" + "test content here"
	if size != expected {
		t.Errorf("expected %d bytes, got %d", expected, size)
	}
}

func TestPathExists(t *testing.T) {
	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(f, []byte("test"), 0644)

	if !PathExists(f) {
		t.Error("existing file should return true")
	}
	if PathExists(filepath.Join(tmpDir, "nonexistent")) {
		t.Error("nonexistent file should return false")
	}
}

func TestIsDir(t *testing.T) {
	tmpDir := t.TempDir()
	if !IsDir(tmpDir) {
		t.Error("directory should return true")
	}

	f := filepath.Join(tmpDir, "file.txt")
	os.WriteFile(f, []byte("test"), 0644)
	if IsDir(f) {
		t.Error("file should return false")
	}
}

func TestExpandHome(t *testing.T) {
	expanded := ExpandHome("~/test")
	if expanded == "~/test" {
		t.Error("expand home should change the path")
	}
}

func TestFindExecutableInPath(t *testing.T) {
	// Should find 'go' since we're in a Go environment
	exe := FindExecutableInPath("go")
	if exe == "" {
		t.Log("go not found in PATH (expected in some environments)")
	} else {
		t.Logf("found go at: %s", exe)
	}
}

func TestFindExecutableInPathPrefersWindowsScriptExtension(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows PATH extension order only applies on Windows")
	}
	tmpDir := t.TempDir()
	name := "devman-test-tool"
	extensionless := filepath.Join(tmpDir, name)
	bat := filepath.Join(tmpDir, name+".bat")
	if err := os.WriteFile(extensionless, []byte("extensionless"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bat, []byte("@echo off\r\n"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", tmpDir)
	found := FindExecutableInPath(name)
	if found != bat {
		t.Fatalf("expected %s, got %s", bat, found)
	}
}

func TestFlutterVersionReadsSdkVersionFile(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "version"), []byte("3.35.7\n"), 0644); err != nil {
		t.Fatal(err)
	}

	got := flutterVersion(tmpDir, "")
	if got != "3.35.7" {
		t.Fatalf("expected Flutter version from SDK version file, got %q", got)
	}
}

func TestIsWindowsAppsAlias(t *testing.T) {
	got := isWindowsAppsAlias(`C:\Users\Administrator\AppData\Local\Microsoft\WindowsApps\python3.exe`)
	if runtime.GOOS == "windows" && !got {
		t.Fatal("expected WindowsApps alias to be detected on Windows")
	}
	if runtime.GOOS != "windows" && got {
		t.Fatal("WindowsApps alias should only be detected on Windows")
	}
}
