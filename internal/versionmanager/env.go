package versionmanager

import (
	"os"
	"os/exec"
)

type Environment interface {
	Getenv(key string) string
	LookPath(command string) string
	DirExists(path string) bool
	FileExists(path string) bool
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

func (RealEnvironment) FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
