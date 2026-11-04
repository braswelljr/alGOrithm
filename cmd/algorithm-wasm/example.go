//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"

	algorithm "github.com/braswelljr/alGOrithm"
)

// example returns the Go source of a program in examples/ ("examples/leaderboard").
func example(args []js.Value) any {
	dir := arg(args, 0)
	code, ok := algorithm.Example(dir)
	if !ok {
		return failure(fmt.Errorf("no example %q", dir))
	}
	return map[string]any{"ok": true, "path": dir, "code": code, "command": "go run ./" + dir}
}
