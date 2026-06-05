package scanner

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// DirSize recursively calculates total size of a directory in bytes
func DirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

// PathExists checks if a path exists
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDir checks if path is a directory
func IsDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// GetEnvPaths returns PATH entries
func GetEnvPaths() []string {
	path := os.Getenv("PATH")
	sep := ":"
	if runtime.GOOS == "windows" {
		sep = ";"
	}
	var entries []string
	for _, p := range strings.Split(path, sep) {
		p = strings.TrimSpace(p)
		if p != "" {
			entries = append(entries, p)
		}
	}
	return entries
}

// ExpandHome expands ~ to home directory
func ExpandHome(path string) string {
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, strings.TrimPrefix(path, "~"))
	}
	return path
}

// ReadFileLines reads lines from a file
func ReadFileLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return strings.Split(string(data), "\n"), nil
}

// FindExecutableInPath looks for an executable in PATH
func FindExecutableInPath(name string) string {
	for _, dir := range GetEnvPaths() {
		full := filepath.Join(dir, name)
		if runtime.GOOS == "windows" {
			exts := []string{".exe", ".cmd", ".bat", ""}
			if filepath.Ext(name) != "" {
				exts = []string{""}
			}
			for _, ext := range exts {
				candidate := full + ext
				if PathExists(candidate) {
					return candidate
				}
			}
		} else {
			if PathExists(full) {
				return full
			}
		}
	}
	return ""
}

func ExecutableVersion(exe string, args ...string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, exe, args...)
	hideCommandWindow(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "unknown"
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return "unknown"
}

// CommonWindowsPaths returns standard installation paths on Windows
func CommonWindowsPaths() []string {
	home, _ := os.UserHomeDir()
	return []string{
		`C:\Program Files`,
		`C:\Program Files (x86)`,
		`C:\`,
		home,
		filepath.Join(home, "AppData", "Local"),
		filepath.Join(home, "scoop", "apps"),
	}
}

// CommonLinuxPaths returns standard installation paths on Linux
func CommonLinuxPaths() []string {
	home, _ := os.UserHomeDir()
	return []string{
		"/usr/local",
		"/usr",
		"/opt",
		home,
		filepath.Join(home, ".local"),
	}
}
