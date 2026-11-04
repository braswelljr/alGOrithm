//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/braswelljr/alGOrithm/internal/core"
)

// measure is measure(id, sizes, order, seed): sizes a JSON array of numbers
// ("" for the defaults), order a dataset kind ("" for random), seed a number
// (0 for 1). It returns {ok, measurement} with the samples and every fit.
func measure(args []js.Value) any {
	var sizes []int
	if s := arg(args, 1); s != "" {
		if err := json.Unmarshal([]byte(s), &sizes); err != nil {
			return failure(fmt.Errorf("sizes must be a JSON array of numbers: %w", err))
		}
	}
	order := arg(args, 2)
	if order == "" {
		order = "random"
	}
	seed := int64(argInt(args, 3))
	if seed == 0 {
		seed = 1
	}
	m, err := core.MeasureOrder(arg(args, 0), order, sizes, seed)
	if err != nil {
		return failure(err)
	}
	return map[string]any{"ok": true, "measurement": m}
}
