package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

const fileName = "devman.log"

// Init configures logrus to write application logs beside the executable.
func Init() (string, error) {
	logPath := Path()
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return logPath, err
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return logPath, err
	}

	logrus.SetOutput(file)
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000",
		CallerPrettyfier: func(frame *runtime.Frame) (string, string) {
			return shortFunction(frame.Function), fmt.Sprintf("%s:%d", trimWorkingDir(frame.File), frame.Line)
		},
	})

	return logPath, nil
}

func Path() string {
	exe, err := os.Executable()
	if err != nil {
		return fileName
	}
	return filepath.Join(filepath.Dir(exe), fileName)
}

func shortFunction(name string) string {
	lastSlash := strings.LastIndex(name, "/")
	if lastSlash >= 0 {
		name = name[lastSlash+1:]
	}
	return name
}

func trimWorkingDir(path string) string {
	wd, err := os.Getwd()
	if err != nil {
		return filepath.ToSlash(path)
	}
	rel, err := filepath.Rel(wd, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}
