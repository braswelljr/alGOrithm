//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"

	"github.com/braswelljr/alGOrithm/internal/core"
)

// quiz asks count questions about a run of the entry on input (its example
// when input is empty), the same for the same seed.
func quiz(args []js.Value) any {
	e, ok := core.Lookup(arg(args, 0))
	if !ok {
		return failure(fmt.Errorf("no entry %q", arg(args, 0)))
	}
	in := core.Input{}
	if e.Example != nil {
		in = *e.Example
	}
	if s := arg(args, 1); s != "" && s != "{}" {
		var err error
		if in, err = decodeInput(s); err != nil {
			return failure(err)
		}
	}
	qs, err := core.Quiz(e.ID, in, argInt(args, 2), int64(argInt(args, 3)))
	if err != nil {
		return failure(err)
	}
	if qs == nil {
		qs = []core.Question{}
	}
	return map[string]any{"ok": true, "questions": qs}
}
