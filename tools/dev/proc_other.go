//go:build !unix

package main

import "os/exec"

func ownGroup(*exec.Cmd) {}

func stopGroup(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}
