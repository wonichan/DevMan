//go:build !windows

package migrator

import "fmt"

func updateWindowsPath(oldPath, newPath string) error {
	// Stub on Linux
	fmt.Println("[stub] updateWindowsPath:", oldPath, "->", newPath)
	return nil
}

func createWindowsJunction(oldPath, newPath string) error {
	// Stub on Linux
	fmt.Println("[stub] createWindowsJunction:", oldPath, "->", newPath)
	return nil
}
