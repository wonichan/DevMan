package scanner

import (
	"os"
	"path/filepath"
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
