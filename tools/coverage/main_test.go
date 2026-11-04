package main

import (
	"strings"
	"testing"
)

const profile = `mode: atomic
github.com/braswelljr/alGOrithm/pkg/a/a.go:3.1,5.2 4 1
github.com/braswelljr/alGOrithm/pkg/a/a.go:7.1,9.2 1 0
github.com/braswelljr/alGOrithm/pkg/b/b.go:3.1,5.2 2 0
github.com/braswelljr/alGOrithm/internal/x/x.go:3.1,5.2 2 0
`

func TestPackageCoverage(t *testing.T) {
	blocks, err := parseProfile(strings.NewReader(profile))
	if err != nil {
		t.Fatal(err)
	}
	got := packageCoverage(blocks)
	if p := got["github.com/braswelljr/alGOrithm/pkg/a"]; p != 80 {
		t.Errorf("pkg/a: got %.1f%%, want 80%%", p)
	}
	if p := got["github.com/braswelljr/alGOrithm/pkg/b"]; p != 0 {
		t.Errorf("pkg/b: got %.1f%%, want 0%%", p)
	}
}

func TestParseProfileRejectsOtherFiles(t *testing.T) {
	if _, err := parseProfile(strings.NewReader("hello\n")); err == nil {
		t.Error("a file without a mode line was read as a profile")
	}
	if _, err := parseProfile(strings.NewReader("mode: set\nnot a line\n")); err == nil {
		t.Error("a malformed line was accepted")
	}
}

func TestCheck(t *testing.T) {
	blocks, _ := parseProfile(strings.NewReader(profile))
	funcs := parseFuncs(strings.NewReader(`github.com/braswelljr/alGOrithm/pkg/a/a.go:3:	Quick		100.0%
github.com/braswelljr/alGOrithm/pkg/a/a.go:7:	helper		0.0%
github.com/braswelljr/alGOrithm/pkg/b/b.go:3:	Lonely		0.0%
github.com/braswelljr/alGOrithm/internal/x/x.go:3:	Skipped		0.0%
total:							(statements)	60.0%
`))
	bad := check(packageCoverage(blocks), funcs, 75)
	want := []string{
		"pkg/b: 0.0% of statements covered, under 75%",
		"pkg/b/b.go: Lonely is exported and no test runs it",
	}
	if strings.Join(bad, "\n") != strings.Join(want, "\n") {
		t.Errorf("got\n%s\nwant\n%s", strings.Join(bad, "\n"), strings.Join(want, "\n"))
	}
}

func TestExported(t *testing.T) {
	for name, want := range map[string]bool{"Quick": true, "quick": false, "": false, "Élan": true} {
		if exported(name) != want {
			t.Errorf("exported(%q) = %v", name, !want)
		}
	}
}

// TestMergesRepeatedBlocks checks a -coverpkg profile, where each test binary
// writes every block: a block counts once, covered if any binary covered it.
func TestMergesRepeatedBlocks(t *testing.T) {
	blocks, err := parseProfile(strings.NewReader(`mode: atomic
github.com/braswelljr/alGOrithm/pkg/a/a.go:3.1,5.2 4 0
github.com/braswelljr/alGOrithm/pkg/a/a.go:3.1,5.2 4 2
github.com/braswelljr/alGOrithm/pkg/a/a.go:3.1,5.2 4 0
github.com/braswelljr/alGOrithm/pkg/a/a.go:7.1,9.2 4 0
`))
	if err != nil {
		t.Fatal(err)
	}
	if p := packageCoverage(blocks)["github.com/braswelljr/alGOrithm/pkg/a"]; p != 50 {
		t.Errorf("got %.1f%%, want 50%% (the first block covered once, the second never)", p)
	}
}
