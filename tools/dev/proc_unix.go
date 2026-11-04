//go:build unix

package main

import (
	"os/exec"
	"syscall"
)

// ownGroup puts the command in a process group of its own, so stopping it
// also stops the processes it starts (pnpm starts next, next starts workers).
func ownGroup(cmd *exec.Cmd) { cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} }

func stopGroup(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
	}
}
