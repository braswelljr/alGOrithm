package terminal

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/braswelljr/gnux/core"
	"github.com/braswelljr/gnux/core/shell"

	"github.com/braswelljr/alGOrithm/web/runtime/tooling"
)

func streams() (core.Streams, *bytes.Buffer, *bytes.Buffer) {
	var out, errs bytes.Buffer
	return core.Streams{In: strings.NewReader(""), Out: &out, Err: &errs}, &out, &errs
}

func TestCommandsIncludeGnuxAndTheExtras(t *testing.T) {
	all := Commands()
	for _, want := range []string{"lsx", "grepx", "go", "bash", "zstd", "gnux", "make"} {
		if !slices.Contains(all, want) {
			t.Errorf("Commands() lacks %s", want)
		}
	}
}

// TestShellRunsToolsAndExtras runs a line through a shell with Exec: a gnux
// tool, the compression commands, a script and the editor hook.
func TestShellRunsToolsAndExtras(t *testing.T) {
	dir := t.TempDir()
	s, out, errs := streams()
	var opened string
	sh := shell.New(shell.Options{}, Exec(func(p string) { opened = p }), s)
	sh.Set("D", dir)
	script := filepath.Join(dir, "s.sh")
	if err := os.WriteFile(script, []byte("echox one $1\nX=inner\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Makefile"), []byte("all:\n\techox made it\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for line, want := range map[string]string{
		"echox hi | trx a-z A-Z": "HI",
		"bash $D/s.sh two":       "one two",
		"bash -c 'echox three'":  "three",
		"make -s -C $D":          "made it",
	} {
		out.Reset()
		if status := sh.Eval(line); status != 0 || !strings.Contains(out.String(), want) {
			t.Errorf("%q: status %d, out %q, err %q", line, status, out.String(), errs.String())
		}
	}
	// A script's variables stay in its own shell.
	out.Reset()
	sh.Eval("echox [$X]")
	if strings.Contains(out.String(), "inner") {
		t.Error("a script's variable leaked into the terminal")
	}
	if status := sh.Eval("echox data > $D/f && gzip $D/f && gunzip $D/f.gz && catx $D/f"); status != 0 {
		t.Errorf("gzip round trip: status %d, %s", status, errs.String())
	}
	sh.Eval("code $D/s.sh")
	if opened != script {
		t.Errorf("code opened %q, want %q", opened, script)
	}
}

func TestStartersParse(t *testing.T) {
	for name, body := range starters {
		if strings.TrimSpace(body) == "" {
			t.Errorf("%s is empty", name)
		}
	}
	if !strings.Contains(starters["main.go"], "package main") {
		t.Error("main.go is not a program")
	}
	for name, body := range starters {
		for _, d := range tooling.Lint(name, body) {
			if d.Severity == tooling.Error {
				t.Errorf("%s:%d: %s", name, d.Line, d.Message)
			}
		}
	}
}
