package utils

import (
	"devman/internal/models"
	"runtime"

	"github.com/sirupsen/logrus"
)

func GetDiskInfo() ([]models.DiskInfo, error) {
	if runtime.GOOS == "windows" {
		disks, err := getWindowsDiskInfo()
		if err != nil {
			logrus.WithError(err).Error("failed to get Windows disk info")
			return nil, err
		}
		logrus.WithField("disk_count", len(disks)).Info("Windows disk info loaded")
		return disks, nil
	}
	// Linux stub
	logrus.WithField("goos", runtime.GOOS).Info("using stub disk info")
	return []models.DiskInfo{
		{Letter: "/", TotalBytes: 100 * 1024 * 1024 * 1024, FreeBytes: 20 * 1024 * 1024 * 1024},
	}, nil
}
