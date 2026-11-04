package core

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestGoSnippetsCompile writes every entry's Go snippet into a module that
// points at this one and builds them all: a copy button must never hand out
// code that does not compile.
func TestGoSnippetsCompile(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a module")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go command")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	mod := "module snippets\n\ngo 1.27\n\nrequire " + Module + " v0.0.0\n\nreplace " + Module + " => " + root + "\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(mod), 0o600); err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, e := range Entries() {
		s := MakeSnippet(e, *e.Example)
		if !strings.HasPrefix(s.CLI, "algorithm run "+e.ID) {
			t.Errorf("%s: CLI %q", e.ID, s.CLI)
		}
		if s.Go == "" {
			continue
		}
		pkg := filepath.Join(dir, strings.ReplaceAll(e.ID, "-", "_"))
		if err := os.MkdirAll(pkg, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(pkg, "main.go"), []byte(s.Go), 0o600); err != nil {
			t.Fatal(err)
		}
		n++
	}
	cmd := exec.Command(goBin, "vet", "./...")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod", "GOWORK=off")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("snippets do not compile:\n%s", out)
	}
	if n < 60 {
		t.Errorf("only %d entries have Go snippets", n)
	}
}
