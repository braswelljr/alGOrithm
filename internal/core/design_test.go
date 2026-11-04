package core

import (
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/systems/design"
)

// TestDesignStudiesNameRealEntries checks that every case study, its
// related entries and the pattern and foundation lists name entries in the
// registry, and that each study's estimate runs.
func TestDesignStudiesNameRealEntries(t *testing.T) {
	known := func(id string) bool { _, ok := Lookup(id); return ok }
	for _, s := range design.Studies() {
		if !known(s.ID) {
			t.Errorf("study %s is not an entry", s.ID)
		}
		for _, r := range s.Related {
			if !known(r) {
				t.Errorf("study %s: related %s is not an entry", s.ID, r)
			}
		}
		if _, err := Execute("back-of-the-envelope", Input{Options: s.Estimate}, 0); err != nil {
			t.Errorf("study %s: estimate: %v", s.ID, err)
		}
		if len(s.Functional) == 0 || len(s.NonFunctional) == 0 || len(s.Tradeoffs) == 0 || len(s.Failures) == 0 {
			t.Errorf("study %s is missing requirements, trade-offs or failures", s.ID)
		}
	}
	for _, id := range append(append([]string(nil), design.Patterns...), design.Foundations...) {
		if !known(id) {
			t.Errorf("%s is not an entry", id)
		}
	}
}

// TestDesignCaseStudiesDrawAnArchitecture checks every case study's example
// draws an architecture for the visualizer.
func TestDesignCaseStudiesDrawAnArchitecture(t *testing.T) {
	for _, s := range design.Studies() {
		e, _ := Lookup(s.ID)
		res, err := Execute(s.ID, *e.Example, 0)
		if err != nil {
			t.Fatalf("%s: %v", s.ID, err)
		}
		if a := res.Output.Architecture; a == nil || len(a.Components) == 0 || len(a.Frames) == 0 {
			t.Errorf("%s draws no architecture", s.ID)
		}
	}
}

// TestDesignOptionsFitJavaScript keeps every design option's bounds within
// JavaScript's safe integers, so the browser's catalog check accepts them.
func TestDesignOptionsFitJavaScript(t *testing.T) {
	const safe = 1<<53 - 1
	for _, e := range Entries() {
		if e.Category != "design" {
			continue
		}
		for _, o := range e.Options {
			if o.Max > safe || o.Min < -safe || o.Default > o.Max || o.Default < o.Min {
				t.Errorf("%s: option %s bounds %d..%d (default %d)", e.ID, o.Name, o.Min, o.Max, o.Default)
			}
		}
	}
}
