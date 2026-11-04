package goexec

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	algorithm "github.com/braswelljr/alGOrithm"
)

// gopath writes the repository's packages under a GOPATH in a temporary
// directory, as the browser build does in its in-memory filesystem.
func gopath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	root := filepath.Join(dir, "src", filepath.FromSlash(Module), "pkg")
	err := fs.WalkDir(algorithm.Source, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := fs.ReadFile(algorithm.Source, p)
		if err != nil {
			return err
		}
		out := filepath.Join(root, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		return os.WriteFile(out, b, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func write(t *testing.T, src string) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), "main.go")
	if err := os.WriteFile(f, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	return f
}

const quick = `package main

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/sorting"
)

func main() { fmt.Println(sorting.Quick([]int{5, 3, 9, 1})) }
`

// TestArchiveSendsEachPackageOnce checks the request: the program, a go.mod
// replacing the repository with ./alg, and the sorting package and the trace
// package it imports, each as one file, edits included.
func TestArchiveSendsEachPackageOnce(t *testing.T) {
	gp := gopath(t)
	edited := filepath.Join(gp, "src", filepath.FromSlash(Module), "pkg", "algorithms", "sorting", "quick.go")
	src, _ := os.ReadFile(edited)
	if err := os.WriteFile(edited, append(src, []byte("\n// EDITED IN THE BROWSER\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	body, err := Archive([]string{write(t, quick)}, gp)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"-- go.mod --",
		"replace " + Module + " => ./alg",
		"-- alg/pkg/algorithms/sorting/sorting.go --",
		"-- alg/pkg/trace/trace.go --",
		"EDITED IN THE BROWSER",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("the request lacks %q", want)
		}
	}
	if n := strings.Count(body, "\n-- "); n+1 > 20 {
		t.Errorf("%d files, over the Playground's 20", n+1)
	}
	// One file per package: a single package clause for sorting.
	part := body[strings.Index(body, "-- alg/pkg/algorithms/sorting/sorting.go --"):]
	if end := strings.Index(part[4:], "\n-- "); end > 0 {
		part = part[:end+4]
	}
	if c := strings.Count(part, "\npackage sorting"); c != 1 {
		t.Errorf("the merged sorting file has %d package clauses, want 1", c)
	}
}

func TestArchiveWithoutTheRepository(t *testing.T) {
	body, err := Archive([]string{write(t, "package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(1) }\n")}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(body, "go.mod") {
		t.Error("a standard-library program was sent a go.mod")
	}
}

func TestArchiveRejectsALibraryPackage(t *testing.T) {
	if _, err := Archive([]string{write(t, "package sorting\n")}, t.TempDir()); err == nil {
		t.Error("go run took a package that is not main")
	}
}

// TestRunPrintsThePlaygroundsAnswer runs against a fake Playground: output to
// stdout and stderr by kind, the status passed on, compile errors to stderr.
func TestRunPrintsThePlaygroundsAnswer(t *testing.T) {
	answer := map[string]any{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil || r.Form.Get("version") != "2" || r.Form.Get("body") == "" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(answer)
	}))
	defer srv.Close()
	old := Endpoint
	Endpoint = srv.URL
	defer func() { Endpoint = old }()
	main := write(t, "package main\n\nfunc main() {}\n")

	answer = map[string]any{"Errors": "", "Status": 0, "Events": []map[string]string{{"Message": "hi\n", "Kind": "stdout"}, {"Message": "oops\n", "Kind": "stderr"}}}
	var out, errs bytes.Buffer
	if s := Command([]string{"run", main}, t.TempDir(), &out, &errs); s != 0 || out.String() != "hi\n" || !strings.Contains(errs.String(), "oops") {
		t.Errorf("status %d, out %q, err %q", s, out.String(), errs.String())
	}
	if !strings.Contains(errs.String(), "play.golang.org") {
		t.Error("go run did not say the program is sent to the Playground")
	}

	answer = map[string]any{"Errors": "./prog.go:3:1: undefined: x", "Status": 0}
	out.Reset()
	errs.Reset()
	if s := Command([]string{"run", main}, t.TempDir(), &out, &errs); s != 1 || !strings.Contains(errs.String(), "undefined: x") {
		t.Errorf("a compile error: status %d, err %q", s, errs.String())
	}

	answer = map[string]any{"Errors": "", "Status": 3}
	if s := Command([]string{"run", main}, t.TempDir(), &out, &errs); s != 3 {
		t.Errorf("exit status 3 came back as %d", s)
	}
}

func TestUsage(t *testing.T) {
	var out, errs bytes.Buffer
	if s := Command(nil, "", &out, &errs); s != 2 || !strings.Contains(out.String(), "go run") {
		t.Errorf("go with no command: status %d", s)
	}
	if s := Command([]string{"build"}, "", &out, &errs); s != 2 {
		t.Errorf("go build: status %d", s)
	}
	if s := Command([]string{"run"}, "", &out, &errs); s != 2 {
		t.Errorf("go run with no files: status %d", s)
	}
}

// TestLivePlayground runs a program on the real Go Playground: set
// GOEXEC_LIVE=1. It is off by default, since it needs the network.
func TestLivePlayground(t *testing.T) {
	if os.Getenv("GOEXEC_LIVE") != "1" {
		t.Skip("set GOEXEC_LIVE=1 to run against play.golang.org")
	}
	var out, errs bytes.Buffer
	status := Command([]string{"run", write(t, quick)}, gopath(t), &out, &errs)
	if status != 0 || out.String() != "[1 3 5 9]\n" {
		t.Fatalf("status %d, out %q, err %q", status, out.String(), errs.String())
	}
}
