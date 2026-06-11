package main

import (
	"devman/internal/models"
	"devman/internal/utils"

	"github.com/sirupsen/logrus"
)

// GetDiskInfo returns disk usage info.
func (a *App) GetDiskInfo() ([]models.DiskInfo, error) {
	disks, err := utils.GetDiskInfo()
	if err != nil {
		logrus.WithError(err).Error("get disk info failed")
		return nil, err
	}
	logrus.WithField("disk_count", len(disks)).Info("get disk info completed")
	return disks, nil
}
