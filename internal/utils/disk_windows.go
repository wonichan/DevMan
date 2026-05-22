//go:build windows

package utils

import (
	"devman/internal/models"
	"syscall"
	"unsafe"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	procGetLogicalDrives = kernel32.NewProc("GetLogicalDrives")
	procGetDiskFreeSpaceExW = kernel32.NewProc("GetDiskFreeSpaceExW")
)

func getWindowsDiskInfo() ([]models.DiskInfo, error) {
	drives, err := getLogicalDrives()
	if err != nil {
		return nil, err
	}

	var result []models.DiskInfo
	for _, drive := range drives {
		var freeBytes, totalBytes, totalFreeBytes uint64
		ret, _, _ := procGetDiskFreeSpaceExW.Call(
			uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(drive))),
			uintptr(unsafe.Pointer(&freeBytes)),
			uintptr(unsafe.Pointer(&totalBytes)),
			uintptr(unsafe.Pointer(&totalFreeBytes)),
		)
		if ret == 0 {
			continue
		}

		used := totalBytes - freeBytes
		var percent float64
		if totalBytes > 0 {
			percent = float64(used) / float64(totalBytes) * 100
		}

		result = append(result, models.DiskInfo{
			Letter:      drive,
			TotalBytes:  int64(totalBytes),
			FreeBytes:   int64(freeBytes),
			UsedBytes:   int64(used),
			UsedPercent: int(percent),
		})
	}
	return result, nil
}

func getLogicalDrives() ([]string, error) {
	ret, _, _ := procGetLogicalDrives.Call()
	if ret == 0 {
		return nil, syscall.GetLastError()
	}
	mask := uint32(ret)
	var drives []string
	for i := 0; i < 26; i++ {
		if mask&(1<<i) != 0 {
			drives = append(drives, string(rune('A'+i))+":")
		}
	}
	return drives, nil
}
