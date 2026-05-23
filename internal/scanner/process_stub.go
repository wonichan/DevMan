//go:build !windows

package scanner

import "os/exec"

func hideCommandWindow(cmd *exec.Cmd) {}
