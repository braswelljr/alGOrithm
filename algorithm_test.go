package algorithm_test

import (
	"io/fs"
	"path"
	"strings"
	"testing"

	algorithm "github.com/braswelljr/alGOrithm"
)

func TestSourceAndDoc(t *testing.T) {
	code, ok := algorithm.File("algorithms/sorting/quick.go")
	if !ok || !strings.Contains(code, "func Quick[") {
		t.Fatal("sorting/quick.go not embedded")
	}
	if d := algorithm.Doc("algorithms/sorting/quick.go", "Quick"); !strings.HasPrefix(d, "Quick returns a sorted copy") {
		t.Errorf("Doc = %q", d)
	}
	if d := algorithm.Doc("structures/stack.go", "Stack"); !strings.HasPrefix(d, "Stack is last in") {
		t.Errorf("type doc = %q", d)
	}
	if d := algorithm.PackageDoc("scheduling"); !strings.HasPrefix(d, "Package scheduling simulates") {
		t.Errorf("PackageDoc = %q", d)
	}
	if algorithm.Doc("algorithms/sorting/quick.go", "Nope") != "" || algorithm.PackageDoc("nope") != "" {
		t.Error("missing names should give empty docs")
	}
}

func TestMethodDoc(t *testing.T) {
	if d := algorithm.Doc("structures/arrays/js.go", "Array.Filter"); d == "" {
		t.Error("no doc for the method Array.Filter")
	}
	if algorithm.Doc("structures/arrays/js.go", "Array.Nope") != "" {
		t.Error("a missing method has a doc")
	}
}

// TestDir finds a package by its name, and every package name is unique, so
// a name alone always names one directory.
func TestDir(t *testing.T) {
	if d := algorithm.Dir("sorting"); d != "algorithms/sorting" {
		t.Errorf("Dir(sorting) = %q", d)
	}
	if d := algorithm.PackageDoc("algorithms/sorting"); !strings.HasPrefix(d, "Package sorting") {
		t.Errorf("PackageDoc by directory = %q", d)
	}
	seen := map[string]string{}
	_ = fs.WalkDir(algorithm.Source, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".go") {
			return err
		}
		dir := path.Dir(p)
		if other, ok := seen[path.Base(dir)]; ok && other != dir {
			t.Errorf("package name %s is used by %s and %s", path.Base(dir), other, dir)
		}
		seen[path.Base(dir)] = dir
		return nil
	})
}

// TestExample reads an example program's source by its directory.
func TestExample(t *testing.T) {
	for _, dir := range []string{"examples/leaderboard", "leaderboard", "examples/leaderboard/"} {
		if code, ok := algorithm.Example(dir); !ok || !strings.Contains(code, "package main") {
			t.Errorf("Example(%q) = %v", dir, ok)
		}
	}
	for _, dir := range []string{"", "nope", "../pkg", "leaderboard/main.go"} {
		if _, ok := algorithm.Example(dir); ok {
			t.Errorf("Example(%q) found a program", dir)
		}
	}
}
