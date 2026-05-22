//go:build !windows

package utils

import (
	"fmt"
	"devman/internal/models"
)

func getWindowsDiskInfo() ([]models.DiskInfo, error) {
	return nil, fmt.Errorf("not on windows")
}
