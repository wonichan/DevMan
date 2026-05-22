package utils

import (
	"devman/internal/models"
	"runtime"
)

func GetDiskInfo() ([]models.DiskInfo, error) {
	if runtime.GOOS == "windows" {
		return getWindowsDiskInfo()
	}
	// Linux stub
	return []models.DiskInfo{
		{Letter: "/", TotalBytes: 100 * 1024 * 1024 * 1024, FreeBytes: 20 * 1024 * 1024 * 1024},
	}, nil
}
