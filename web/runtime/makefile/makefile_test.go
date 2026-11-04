package makefile

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/braswelljr/gnux/cmd"
	"github.com/braswelljr/gnux/core"
	"github.com/braswelljr/gnux/core/shell"
)

func eval(line string, s core.Streams) int {
	sh := shell.New(shell.Options{}, func(name string, args []string, s core.Streams) (int, bool) {
		return cmd.RunTool(name, args, s)
	}, s)
	return sh.Eval(line)
}

// run writes a Makefile in a fresh directory and runs make there.
func run(t *testing.T, makefile string, args ...string) (int, string, string) {
	t.Helper()
	if makefile != "" {
		if err := os.WriteFile("Makefile", []byte(makefile), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var out, errb bytes.Buffer
	status := Run(args, core.Streams{In: strings.NewReader(""), Out: &out, Err: &errb}, eval)
	return status, out.String(), errb.String()
}

func TestRecipesAndVariables(t *testing.T) {
	t.Chdir(t.TempDir())
	mk := "NAME = world\nGREETING := hello $(NAME)\nX ?= unset\nX += more\n\n.PHONY: all\nall: greet\n\t@echo done $(X)\n\ngreet:\n\techo $(GREETING) from $@\n"
	status, out, errs := run(t, mk)
	if status != 0 {
		t.Fatalf("status %d: %s", status, errs)
	}
	want := "echo hello world from greet\nhello world from greet\ndone unset more\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

func TestCommandLineWins(t *testing.T) {
	t.Chdir(t.TempDir())
	_, out, _ := run(t, "N = 1\nshow:\n\t@echo $(N) $${HOME:+set}\n", "N=2", "-s")
	if out != "2 set\n" {
		t.Errorf("out = %q", out)
	}
}

func TestUpToDate(t *testing.T) {
	t.Chdir(t.TempDir())
	mk := "out.txt: in.txt\n\tcatx $< > $@\n"
	if err := os.WriteFile("in.txt", []byte("data\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if status, _, errs := run(t, mk); status != 0 {
		t.Fatalf("first build: %s", errs)
	}
	if b, _ := os.ReadFile("out.txt"); string(b) != "data\n" {
		t.Fatalf("out.txt = %q", b)
	}
	_, out, _ := run(t, "")
	if !strings.Contains(out, "'out.txt' is up to date") {
		t.Errorf("second build: %q", out)
	}
	future := time.Now().Add(time.Hour)
	if err := os.Chtimes("in.txt", future, future); err != nil {
		t.Fatal(err)
	}
	if _, out, _ := run(t, ""); !strings.Contains(out, "catx in.txt > out.txt") {
		t.Errorf("after a change: %q", out)
	}
}

func TestFailures(t *testing.T) {
	t.Chdir(t.TempDir())
	status, _, errs := run(t, "a:\n\t-falsex\n\tfalsex\n\techo never\n")
	if status != 2 || !strings.Contains(errs, "(ignored)") || !strings.Contains(errs, "[Makefile:3: a] Error 1") {
		t.Errorf("status %d, errs %q", status, errs)
	}
	if status, _, errs := run(t, "a: missing\n\techo\n"); status != 2 || !strings.Contains(errs, "needed by 'a'") {
		t.Errorf("a missing prerequisite: %d %q", status, errs)
	}
	if status, _, errs := run(t, "x\n"); status != 2 || !strings.Contains(errs, "Makefile:1: *** missing separator") {
		t.Errorf("a bad line: %d %q", status, errs)
	}
	if status, _, errs := run(t, "\techo\n"); status != 2 || !strings.Contains(errs, "recipe commences before first target") {
		t.Errorf("an early recipe: %d %q", status, errs)
	}
}

func TestDryRunDirectoryAndFile(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.Mkdir("sub", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("sub/build.mk", []byte("go: ; touchx made\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	status, out, errs := run(t, "", "-n", "-C", "sub", "-f", "build.mk")
	if status != 0 || !strings.Contains(out, "touchx made") || !strings.Contains(out, "Entering directory") {
		t.Fatalf("%d %q %q", status, out, errs)
	}
	if _, err := os.Stat("sub/made"); err == nil {
		t.Error("-n ran the recipe")
	}
	if wd, _ := os.Getwd(); wd != dir {
		t.Errorf("-C left the directory at %s", wd)
	}
	if status, _, _ := run(t, "", "nothing-here"); status != 2 {
		t.Errorf("no makefile: %d", status)
	}
}

func TestCycleAndNothingToDo(t *testing.T) {
	t.Chdir(t.TempDir())
	_, out, errs := run(t, ".PHONY: a b\na: b\nb: a\n")
	if !strings.Contains(errs, "Circular") || !strings.Contains(out, "Nothing to be done for 'a'") {
		t.Errorf("out %q errs %q", out, errs)
	}
}
