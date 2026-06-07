package versionmanager

import (
	"os"
	"os/exec"
	"path/filepath"
)

type Environment interface {
	Getenv(key string) string
	LookPath(command string) string
	DirExists(path string) bool
}

type MutableEnvironment interface {
	Environment
	ExecutableDir() string
	WriteFile(path string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	SetUserEnv(key string, value string) error
	EnsureUserPathEntry(entry string) error
	Run(command string, args ...string) (string, error)
}

type RealEnvironment struct{}

func (RealEnvironment) Getenv(key string) string {
	return os.Getenv(key)
}

func (RealEnvironment) LookPath(command string) string {
	path, err := exec.LookPath(command)
	if err != nil {
		return ""
	}
	return path
}

func (RealEnvironment) DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func (RealEnvironment) ExecutableDir() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(exe)
}

func (RealEnvironment) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

func (RealEnvironment) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (RealEnvironment) SetUserEnv(key string, value string) error {
	return setUserEnv(key, value)
}

func (RealEnvironment) EnsureUserPathEntry(entry string) error {
	return ensureUserPathEntry(entry)
}

func (RealEnvironment) Run(command string, args ...string) (string, error) {
	output, err := exec.Command(command, args...).CombinedOutput()
	return string(output), err
}
