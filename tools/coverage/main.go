// Command coverage checks a coverage profile against the floor the
// repository keeps: every package under pkg/ at or above a statement
// coverage (90% by default), and no exported function in pkg/ that no test
// runs at all. CI runs it after the tests (make cover/check).
//
//	go test -coverprofile=coverage.out ./...
//	go run ./tools/coverage -profile coverage.out -min 90
//
// It exits 1 listing every package under the floor and every untested
// exported function, and 2 on a usage mistake.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// module is the import path every profile line starts with.
const module = "github.com/braswelljr/alGOrithm/"

// block is one line of a coverage profile: a run of statements and how often they ran.
type block struct {
	file  string
	at    string // the position, "12.3,14.2"
	stmts int
	count int
}

// parseProfile reads a profile written by go test -coverprofile.
func parseProfile(r io.Reader) ([]block, error) {
	var out []block
	sc := bufio.NewScanner(r)
	first := true
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if first {
			first = false
			if !strings.HasPrefix(line, "mode:") {
				return nil, fmt.Errorf("not a coverage profile: %q", line)
			}
			continue
		}
		if line == "" {
			continue
		}
		// file.go:12.3,14.2 3 1
		colon := strings.LastIndex(line, ":")
		fields := strings.Fields(line[colon+1:])
		if colon < 0 || len(fields) != 3 {
			return nil, fmt.Errorf("bad profile line: %q", line)
		}
		stmts, err1 := strconv.Atoi(fields[1])
		count, err2 := strconv.Atoi(fields[2])
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("bad profile line: %q", line)
		}
		out = append(out, block{file: line[:colon], at: fields[0], stmts: stmts, count: count})
	}
	return out, sc.Err()
}

// packageCoverage is each package's covered share of its statements, in
// percent. A profile written with -coverpkg holds each block once per test
// binary; a block counts once, and as covered if any binary ran it.
func packageCoverage(blocks []block) map[string]float64 {
	type key struct{ file, at string }
	merged := map[key]block{}
	for _, b := range blocks {
		k := key{b.file, b.at}
		if m, ok := merged[k]; ok && m.count >= b.count {
			continue
		}
		merged[k] = b
	}
	total := map[string]int{}
	covered := map[string]int{}
	for _, b := range merged {
		pkg := path.Dir(b.file)
		total[pkg] += b.stmts
		if b.count > 0 {
			covered[pkg] += b.stmts
		}
	}
	out := map[string]float64{}
	for pkg, n := range total {
		if n > 0 {
			out[pkg] = 100 * float64(covered[pkg]) / float64(n)
		}
	}
	return out
}

// fn is one line of go tool cover -func: a function and its coverage.
type fn struct {
	file string
	line int
	name string
	pct  float64
}

// parseFuncs reads the output of go tool cover -func.
func parseFuncs(r io.Reader) []fn {
	var out []fn
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) != 3 || !strings.HasSuffix(f[2], "%") {
			continue
		}
		pct, err := strconv.ParseFloat(strings.TrimSuffix(f[2], "%"), 64)
		if err != nil {
			continue
		}
		file, line := f[0], 0
		if i := strings.Index(file, ".go:"); i >= 0 {
			line, _ = strconv.Atoi(strings.TrimSuffix(file[i+4:], ":"))
			file = file[:i+3]
		}
		out = append(out, fn{file: file, line: line, name: f[1], pct: pct})
	}
	return out
}

// exported reports whether a function or method name is exported.
func exported(name string) bool {
	r := []rune(name)
	return len(r) > 0 && unicode.IsUpper(r[0])
}

// inLibrary reports whether a file is in the library, pkg/.
func inLibrary(file string) bool {
	return strings.HasPrefix(strings.TrimPrefix(file, module), "pkg/")
}

// empty reports whether the function declared at file:line has no statements
// (a no-op such as trace.Nop's Record): cover reports it at 0% whatever runs it.
// A file it cannot read counts as not empty, so a doubt is reported.
func empty(f fn) bool {
	src, err := os.ReadFile(strings.TrimPrefix(f.file, module))
	if err != nil {
		return false
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, f.file, src, 0)
	if err != nil {
		return false
	}
	for _, d := range file.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fset.Position(fd.Pos()).Line == f.line {
			return fd.Body != nil && len(fd.Body.List) == 0
		}
	}
	return false
}

// check returns the failures: packages under min, and exported functions no test runs.
func check(pkgs map[string]float64, funcs []fn, floor float64) []string {
	var bad []string
	names := make([]string, 0, len(pkgs))
	for p := range pkgs {
		names = append(names, p)
	}
	sort.Strings(names)
	for _, p := range names {
		if inLibrary(p+"/") && pkgs[p] < floor {
			bad = append(bad, fmt.Sprintf("%s: %.1f%% of statements covered, under %.0f%%", strings.TrimPrefix(p, module), pkgs[p], floor))
		}
	}
	for _, f := range funcs {
		if inLibrary(f.file) && exported(f.name) && f.pct == 0 && !empty(f) {
			bad = append(bad, fmt.Sprintf("%s: %s is exported and no test runs it", strings.TrimPrefix(f.file, module), f.name))
		}
	}
	return bad
}

func main() {
	profile := flag.String("profile", "coverage.out", "the coverage profile")
	floor := flag.Float64("min", 90, "the least statement coverage a pkg/ package may have, in percent")
	flag.Parse()

	file, err := os.Open(*profile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "coverage:", err)
		os.Exit(2)
	}
	blocks, err := parseProfile(file)
	_ = file.Close() // read only: a failed close loses nothing
	if err != nil {
		fmt.Fprintln(os.Stderr, "coverage:", err)
		os.Exit(2)
	}
	out, err := exec.Command("go", "tool", "cover", "-func="+*profile).Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, "coverage: go tool cover:", err)
		os.Exit(2)
	}
	bad := check(packageCoverage(blocks), parseFuncs(strings.NewReader(string(out))), *floor)
	if len(bad) > 0 {
		fmt.Fprintf(os.Stderr, "coverage: %d problems\n", len(bad))
		for _, b := range bad {
			fmt.Fprintln(os.Stderr, "  "+b)
		}
		os.Exit(1)
	}
	fmt.Printf("coverage: every pkg/ package at %.0f%% or more, every exported function tested\n", *floor)
}
