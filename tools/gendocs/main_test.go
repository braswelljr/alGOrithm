package main

import (
	"bytes"
	"os"
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
