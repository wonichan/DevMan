package main

import (
	"devman/internal/logger"
	"embed"

	"github.com/sirupsen/logrus"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	logPath, err := logger.Init()
	if err != nil {
		println("Failed to initialize logger:", err.Error())
	} else {
		logrus.WithField("path", logPath).Info("logger initialized")
	}

	app := NewApp()

	err = wails.Run(&options.App{
		Title:     "DevMan",
		Width:     1280,
		Height:    800,
		MinWidth:  980,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 15, G: 23, B: 42, A: 255},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		logrus.WithError(err).Error("wails application exited with error")
		println("Error:", err.Error())
	}
}
