package tooling

import (
	"errors"
	"strings"
	"testing"
)

func TestKindOf(t *testing.T) {
	for name, want := range map[string]Kind{
		"main.go": Go, "a/Makefile": Make, "rules.mk": Make, "x.json": JSON,
		"run.sh": Shell, "notes.md": Text, "GNUmakefile": Make,
	} {
		if got := KindOf(name); got != want {
			t.Errorf("KindOf(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestFormatGo(t *testing.T) {
	got, err := Format("main.go", "package main\nfunc main(){\nprintln( 1 )}\n")
	if err != nil {
		t.Fatal(err)
	}
	want := "package main\n\nfunc main() {\n\tprintln(1)\n}\n"
	if got != want {
		t.Errorf("Format = %q, want %q", got, want)
	}
	if _, err := Format("main.go", "package main\nfunc {"); !errors.Is(err, ErrSyntax) {
		t.Errorf("Format of a broken file: err = %v, want ErrSyntax", err)
	}
}

func TestFormatOthers(t *testing.T) {
	if got, _ := Format("x.json", `{"a":[1,2]}`); got != "{\n  \"a\": [\n    1,\n    2\n  ]\n}\n" {
		t.Errorf("json: %q", got)
	}
	if got, _ := Format("run.sh", "echo hi   \n\n\n"); got != "echo hi\n" {
		t.Errorf("shell: %q", got)
	}
	if got, _ := Format("Makefile", "all:\n\techo hi  "); got != "all:\n\techo hi\n" {
		t.Errorf("make keeps tabs: %q", got)
	}
	if got, _ := Format("empty.txt", "\n\n"); got != "" {
		t.Errorf("empty: %q", got)
	}
}

func messages(ds []Diagnostic) string {
	var b strings.Builder
	for _, d := range ds {
		b.WriteString(d.Message)
		b.WriteString("\n")
	}
	return b.String()
}

func TestLintGo(t *testing.T) {
	src := "package main\n\nimport (\n\t\"fmt\"\n\t\"os\"\n)\n\nfunc main() {\n\tx := 1\n\tfmt.Println(y)\n}\n"
	ds := Lint("main.go", src)
	all := messages(ds)
	for _, want := range []string{`"os" imported and not used`, "declared and not used: x", "undefined: y"} {
		if !strings.Contains(all, want) {
			t.Errorf("Lint is missing %q; got:\n%s", want, all)
		}
	}
	for _, d := range ds {
		if d.Line < 1 || d.Col < 1 {
			t.Errorf("bad position %+v", d)
		}
	}
	// What is used from an import is not reported: the import is unknown here.
	if ds := Lint("ok.go", "package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(fmt.Sprint(1) + 2) }\n"); len(ds) != 0 {
		t.Errorf("a well-typed file: %+v", ds)
	}
}

func TestUnusedImportNames(t *testing.T) {
	src := "package main\n\nimport (\n\t\"math/rand/v2\"\n\tr \"strings\"\n\t_ \"embed\"\n)\n\nfunc main() { _ = rand.N(3); _ = r.ToUpper(\"a\") }\n"
	for _, d := range Lint("main.go", src) {
		if d.Severity == Error {
			t.Errorf("Lint: %+v", d)
		}
	}
}

func TestLintGoSyntax(t *testing.T) {
	ds := Lint("main.go", "package main\nfunc main() {\n")
	if len(ds) == 0 || ds[0].Severity != Error || ds[0].Source != "go" {
		t.Fatalf("Lint of a broken file = %+v", ds)
	}
}

func TestLintJSON(t *testing.T) {
	ds := Lint("x.json", "{\n  \"a\": 1,\n}\n")
	if len(ds) != 1 || ds[0].Line != 3 || ds[0].Severity != Error {
		t.Errorf("Lint = %+v", ds)
	}
	if ds := Lint("x.json", "{}\n"); len(ds) != 0 {
		t.Errorf("valid json: %+v", ds)
	}
}

func TestLintMake(t *testing.T) {
	ds := Lint("Makefile", "CC := go\n\nbuild: dep\n    go build\n\tgo vet\n")
	if len(ds) != 1 || ds[0].Line != 4 || ds[0].Source != "make" {
		t.Errorf("Lint = %+v", ds)
	}
	if ds := Lint("Makefile", "\techo early\n"); len(ds) != 1 {
		t.Errorf("a recipe before a rule: %+v", ds)
	}
	if ds := Lint("Makefile", "X = a:b\n  Y = 1\n"); len(ds) != 0 {
		t.Errorf("assignments are not rules: %+v", ds)
	}
}

func TestLintShell(t *testing.T) {
	if ds := Lint("run.sh", "if true; then\n  echo hi\n"); len(ds) != 1 || ds[0].Line != 2 {
		t.Errorf("an open if: %+v", ds)
	}
	if ds := Lint("run.sh", "echo hi\n"); len(ds) != 0 {
		t.Errorf("a finished script: %+v", ds)
	}
}

func TestStyle(t *testing.T) {
	ds := Lint("notes.txt", "a  \nb")
	if len(ds) != 2 || ds[0].Col != 2 || ds[0].EndCol != 4 || ds[1].Line != 2 {
		t.Errorf("Lint = %+v", ds)
	}
}
