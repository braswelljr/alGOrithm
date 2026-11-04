// Package goexec is the terminal's `go` command. `go run` collects the program
// and every package of this repository it imports, read from source under
// $GOPATH/src so an edited algorithm runs as edited, and runs them on the Go
// Playground (play.golang.org), which compiles real Go. It says so before it
// sends anything: the Playground is a Google service, and this is the one
// command in the terminal that leaves the browser.
//
// The Playground takes at most 20 files, so each package goes as one file:
// its files' import blocks merged and their declarations joined, which Go
// reads as the same package.
package goexec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Module is this repository's module path: imports under it are sent from
// source, everything else is the standard library.
const Module = "github.com/braswelljr/alGOrithm"

// Endpoint is where programs are compiled and run.
var Endpoint = "https://play.golang.org/compile"

// Client sends the requests; tests swap in a fake Playground.
var Client = &http.Client{Timeout: 60 * time.Second}

// Command runs `go ARGS...` with gopath as GOPATH, writing to the given
// streams, and returns the exit status: 0 success, 1 a failed program, 2 a
// usage mistake.
func Command(args []string, gopath string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprint(stdout, usage)
		if len(args) == 0 {
			return 2
		}
		return 0
	}
	switch args[0] {
	case "version":
		fmt.Fprintln(stdout, "go (runs on the Go Playground, play.golang.org)")
		return 0
	case "run":
		files, rest := splitRun(args[1:])
		if len(files) == 0 {
			fmt.Fprintln(stderr, "go run: no go files listed")
			return 2
		}
		if len(rest) > 0 {
			fmt.Fprintln(stderr, "go run: the Go Playground takes no program arguments")
			return 2
		}
		return Run(files, gopath, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "go %s: unknown command\nRun 'go help' for usage.\n", args[0])
		return 2
	}
}

const usage = `Go runs Go programs on the Go Playground (play.golang.org): real Go,
compiled and run by Google's service. The program and the packages of this
repository it imports are sent there; nothing else in this terminal is.

Usage:

	go run FILE.go [FILE.go...]   run a program
	go version                    say where programs run
	go help                       this text

The repository's packages (github.com/braswelljr/alGOrithm/pkg/...) are read
from source under $GOPATH/src, so an edited algorithm runs as edited.
`

func splitRun(args []string) (files, rest []string) {
	for i, a := range args {
		if !strings.HasSuffix(a, ".go") {
			return files, args[i:]
		}
		files = append(files, a)
	}
	return files, nil
}

// Run builds the request for files, sends it, and prints what the program
// printed, its compile errors, or why it could not be sent.
func Run(files []string, gopath string, stdout, stderr io.Writer) int {
	body, err := Archive(files, gopath)
	if err != nil {
		fmt.Fprintf(stderr, "go run: %v\n", err)
		return 1
	}
	fmt.Fprintln(stderr, "go run: compiling on the Go Playground (play.golang.org); the program is sent there")
	res, err := send(body)
	if err != nil {
		fmt.Fprintf(stderr, "go run: %v\n", err)
		return 1
	}
	if res.Errors != "" {
		fmt.Fprint(stderr, res.Errors)
		if !strings.HasSuffix(res.Errors, "\n") {
			fmt.Fprintln(stderr)
		}
		return 1
	}
	for _, e := range res.Events {
		if e.Kind == "stderr" {
			fmt.Fprint(stderr, e.Message)
		} else {
			fmt.Fprint(stdout, e.Message)
		}
	}
	return res.Status
}

// result is the Playground's answer.
type result struct {
	Errors string
	Events []struct {
		Message string
		Kind    string
	}
	Status int
}

func send(body string) (result, error) {
	var r result
	resp, err := Client.PostForm(Endpoint, url.Values{"body": {body}, "version": {"2"}})
	if err != nil {
		return r, fmt.Errorf("the Go Playground could not be reached: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return r, fmt.Errorf("the Go Playground answered %s", resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return r, fmt.Errorf("the Go Playground's answer could not be read: %w", err)
	}
	return r, nil
}

// Archive is the Playground request for the program in files: the program,
// a go.mod replacing this repository with a local copy, and one merged file
// for each repository package the program imports, directly or not.
func Archive(files []string, gopath string) (string, error) {
	var main bytes.Buffer
	var needs []string
	for i, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			return "", err
		}
		imports, body, pkg, err := split(f, src)
		if err != nil {
			return "", err
		}
		if pkg != "main" {
			return "", fmt.Errorf("%s is package %s; go run needs package main", f, pkg)
		}
		if i == 0 {
			main.Write(src)
		} else {
			// Later files of the program go as their own files.
			_ = body
		}
		needs = append(needs, local(imports)...)
	}
	parts := []string{main.String()}
	for _, f := range files[1:] {
		src, _ := os.ReadFile(f)
		parts = append(parts, "-- "+filepath.Base(f)+" --\n"+string(src))
	}
	pkgs, err := closure(needs, gopath)
	if err != nil {
		return "", err
	}
	if len(pkgs) > 0 {
		parts = append(parts,
			"-- go.mod --\nmodule play\n\ngo 1.27\n\nrequire "+Module+" v0.0.0\n\nreplace "+Module+" => ./alg\n",
			"-- alg/go.mod --\nmodule "+Module+"\n\ngo 1.27\n")
		for _, p := range pkgs {
			merged, err := mergePackage(dirOf(p, gopath))
			if err != nil {
				return "", err
			}
			rel := strings.TrimPrefix(p, Module+"/")
			parts = append(parts, "-- alg/"+rel+"/"+path.Base(rel)+".go --\n"+merged)
		}
	}
	if len(parts)+2 > 20 {
		return "", fmt.Errorf("the program needs %d files; the Go Playground takes 20", len(parts)+2)
	}
	return strings.Join(parts, "\n"), nil
}

// local keeps the imports that are this repository's packages.
func local(imports []string) []string {
	var out []string
	for _, p := range imports {
		if strings.HasPrefix(p, Module+"/") {
			out = append(out, p)
		}
	}
	return out
}

func dirOf(pkg, gopath string) string {
	return filepath.Join(gopath, "src", filepath.FromSlash(pkg))
}

// closure is every repository package reachable from roots, sorted.
func closure(roots []string, gopath string) ([]string, error) {
	seen := map[string]bool{}
	var walk func(p string) error
	walk = func(p string) error {
		if seen[p] {
			return nil
		}
		seen[p] = true
		files, err := goFiles(dirOf(p, gopath))
		if err != nil {
			return fmt.Errorf("package %s: %w", p, err)
		}
		for _, f := range files {
			src, err := os.ReadFile(f)
			if err != nil {
				return err
			}
			imports, _, _, err := split(f, src)
			if err != nil {
				return err
			}
			for _, q := range local(imports) {
				if err := walk(q); err != nil {
					return err
				}
			}
		}
		return nil
	}
	for _, r := range roots {
		if err := walk(r); err != nil {
			return nil, err
		}
	}
	out := make([]string, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	sort.Strings(out)
	return out, nil
}

// goFiles lists a package's source files: no tests, and none limited by a
// build constraint (the Playground builds for one platform; the library's
// constrained files are for other ones).
func goFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		n := e.Name()
		if e.IsDir() || !strings.HasSuffix(n, ".go") || strings.HasSuffix(n, "_test.go") {
			continue
		}
		src, err := os.ReadFile(filepath.Join(dir, n))
		if err != nil {
			return nil, err
		}
		if constrained(src) {
			continue
		}
		out = append(out, filepath.Join(dir, n))
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no Go files in %s", dir)
	}
	return out, nil
}

func constrained(src []byte) bool {
	for _, line := range strings.Split(string(src), "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "//go:build") {
			return true
		}
		if strings.HasPrefix(t, "package ") {
			return false
		}
	}
	return false
}

// split reads a file's imports, the source after its import declarations,
// and its package name.
func split(name string, src []byte) (imports []string, body []byte, pkg string, err error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, name, src, parser.ImportsOnly|parser.ParseComments)
	if err != nil {
		return nil, nil, "", err
	}
	end := fset.Position(f.Name.End()).Offset
	for _, d := range f.Decls {
		if e := fset.Position(d.End()).Offset; e > end {
			end = e
		}
	}
	for _, s := range f.Imports {
		p, _ := strconv.Unquote(s.Path.Value)
		imports = append(imports, p)
	}
	return imports, src[end:], f.Name.Name, nil
}

// mergePackage is a package's files as one: its package clause, the union of
// their imports (with their names, if any), and each file's declarations.
func mergePackage(dir string) (string, error) {
	files, err := goFiles(dir)
	if err != nil {
		return "", err
	}
	specs := map[string]bool{}
	var pkg string
	var bodies []string
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			return "", err
		}
		fset := token.NewFileSet()
		pf, err := parser.ParseFile(fset, f, src, parser.ImportsOnly)
		if err != nil {
			return "", err
		}
		pkg = pf.Name.Name
		for _, s := range pf.Imports {
			spec := s.Path.Value
			if s.Name != nil {
				spec = s.Name.Name + " " + spec
			}
			specs[spec] = true
		}
		_, body, _, err := split(f, src)
		if err != nil {
			return "", err
		}
		bodies = append(bodies, "// "+filepath.Base(f)+"\n"+strings.TrimLeft(string(body), "\n"))
	}
	list := make([]string, 0, len(specs))
	for s := range specs {
		list = append(list, "\t"+s)
	}
	sort.Strings(list)
	var b strings.Builder
	fmt.Fprintf(&b, "package %s\n\n", pkg)
	if len(list) > 0 {
		fmt.Fprintf(&b, "import (\n%s\n)\n\n", strings.Join(list, "\n"))
	}
	b.WriteString(strings.Join(bodies, "\n"))
	return b.String(), nil
}
