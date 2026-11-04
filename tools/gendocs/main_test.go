package main

import (
	"bytes"
	"os"
	"regexp"
	"strings"
	"testing"
)

// TestGeneratedDocsAreCurrent fails when README.md or docs/catalog.md
// no longer match the registry: run `go run ./tools/gendocs`.
func TestGeneratedDocsAreCurrent(t *testing.T) {
	if err := os.Chdir("../.."); err != nil {
		t.Fatal(err)
	}
	for path, want := range files() {
		got, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(got, want) {
			t.Errorf("%s is out of date; run `go run ./tools/gendocs`", path)
		}
	}
}

func TestSplice(t *testing.T) {
	doc := "a\n" + begin + "\nold\n" + end + "\nz"
	if got := Splice(doc, "new\n"); got != "a\n"+begin+"\nnew\n"+end+"\nz" {
		t.Errorf("Splice = %q", got)
	}
	if Splice("no markers", "x") != "no markers" {
		t.Error("Splice changed a document without markers")
	}
}

// TestLearnLinksLandOnHeadings checks docs/learn.md is plain Markdown: no HTML
// anchors, and every in-page link names a heading's own anchor.
func TestLearnLinksLandOnHeadings(t *testing.T) {
	doc := Learn()
	if strings.Contains(doc, "<a ") {
		t.Error("learn.md has an HTML anchor")
	}
	headings := map[string]bool{}
	for _, line := range strings.Split(doc, "\n") {
		if h, ok := strings.CutPrefix(line, "## "); ok {
			headings[anchor(h)] = true
		}
	}
	for _, m := range regexp.MustCompile(`\]\(#([^)]+)\)`).FindAllStringSubmatch(doc, -1) {
		if !headings[m[1]] {
			t.Errorf("link #%s has no heading", m[1])
		}
	}
	if got := anchor("Big O by example: a, b & c"); got != "big-o-by-example-a-b--c" {
		t.Errorf("anchor = %q", got)
	}
}
