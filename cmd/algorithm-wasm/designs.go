//go:build js && wasm

package main

import "github.com/braswelljr/alGOrithm/pkg/systems/design"

// designs is the system design case studies, and the pattern and foundation
// entries, for the site's system design hub.
func designs() any {
	return map[string]any{"ok": true, "studies": design.Studies(), "patterns": design.Patterns, "foundations": design.Foundations}
}
