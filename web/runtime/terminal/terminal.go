// Package terminal is the editor's terminal: gnux's shell with every gnux
// tool, plus the commands a Go playground needs. It is plain Go, so it is
// tested natively; cmd/terminal wires it to the browser.
//
//	go run FILE.go      run a program (on the Go Playground; see goexec)
//	bash FILE, sh FILE  run a shell script in a shell of its own
//	make [TARGET]       run a Makefile's recipes (see package makefile)
//	gzip, gunzip        compress with gzip (klauspost/compress)
//	zstd, unzstd        compress with Zstandard (klauspost/compress)
//	gnux ...            gnux itself: gnux --list, gnux help rmx
//	every gnux tool     lsx, catx, grepx, sedx, findx, ... (and the plain
//	                    GNU names: ls, cat, grep, ...)
package terminal

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	algorithm "github.com/braswelljr/alGOrithm"
	"github.com/braswelljr/gnux/cmd"
	"github.com/braswelljr/gnux/core"
	"github.com/braswelljr/gnux/core/shell"

	"github.com/braswelljr/alGOrithm/web/runtime/goexec"
	"github.com/braswelljr/alGOrithm/web/runtime/makefile"
	"github.com/braswelljr/alGOrithm/web/runtime/ztools"
)

// Home is the visitor's home directory in the terminal's filesystem.
const Home = "/home/guest"

// GoPath is where `go run` reads this repository's packages from.
const GoPath = Home + "/go"

// Library is where the repository's packages are, for the editor's sidebar.
const Library = GoPath + "/src/" + goexec.Module + "/pkg"

// Extra are the commands added to gnux's tools, with a line each for help.
var Extra = map[string]string{
	"go":     "run a Go program (on the Go Playground): go run FILE.go",
	"bash":   "run a shell script: bash FILE",
	"sh":     "run a shell script: sh FILE",
	"make":   "run a Makefile's recipes: make [TARGET]",
	"gzip":   "compress files with gzip",
	"gunzip": "decompress .gz files",
	"zstd":   "compress files with Zstandard",
	"unzstd": "decompress .zst files",
	"gnux":   "gnux itself: gnux --list, gnux help TOOL",
}

// Commands is every command the terminal knows, sorted: gnux's tools and Extra.
func Commands() []string {
	seen := map[string]bool{}
	for _, t := range cmd.Tools() {
		seen[t.Name()] = true
	}
	for name := range Extra {
		seen[name] = true
	}
	out := make([]string, 0, len(seen))
	for n := range seen {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// Exec is the shell's command runner: the extras first, then gnux's tools.
// Open, when set, is how `code FILE` asks the editor to open a file.
func Exec(open func(path string)) shell.ExecFunc {
	var self shell.ExecFunc
	self = func(name string, args []string, s core.Streams) (int, bool) {
		switch name {
		case "go":
			return goexec.Command(args, GoPath, s.Out, s.Err), true
		case "gzip", "gunzip", "zstd", "unzstd":
			return ztools.Run(name, args, s.In, s.Out, s.Err), true
		case "bash", "sh":
			return script(name, args, s, self), true
		case "make":
			// Each recipe line in a shell of its own, as make runs them.
			return makefile.Run(args, s, func(line string, rs core.Streams) int {
				sh := shell.New(shell.Options{}, self, rs)
				sh.LoopLimit = 10000
				return sh.Eval(line)
			}), true
		case "code":
			if open == nil {
				fmt.Fprintln(s.Err, "code: there is no editor here")
				return 1, true
			}
			if len(args) != 1 {
				fmt.Fprintln(s.Err, "usage: code FILE")
				return 2, true
			}
			p, err := filepath.Abs(args[0])
			if err != nil {
				fmt.Fprintf(s.Err, "code: %v\n", err)
				return 1, true
			}
			open(p)
			return 0, true
		case "gnux":
			root := cmd.NewRoot(s)
			root.SetArgs(args)
			if err := root.Execute(); err != nil {
				return 1, true
			}
			return 0, true
		}
		return cmd.RunTool(name, args, s)
	}
	return self
}

// script runs a shell script in a shell of its own, as `bash FILE` does: its
// variables and cd stay in it. -c runs the next argument as the script.
func script(name string, args []string, s core.Streams, exec shell.ExecFunc) int {
	if len(args) == 0 {
		fmt.Fprintf(s.Err, "%s: give a script, or -c COMMANDS\n", name)
		return 2
	}
	var text string
	rest := args[1:]
	if args[0] == "-c" {
		if len(args) < 2 {
			fmt.Fprintf(s.Err, "%s: -c: option requires an argument\n", name)
			return 2
		}
		text, rest = args[1], args[2:]
	} else {
		b, err := os.ReadFile(args[0])
		if err != nil {
			fmt.Fprintf(s.Err, "%s: %v\n", name, err)
			return 1
		}
		text = string(b)
	}
	sh := shell.New(shell.Options{}, exec, s)
	sh.LoopLimit = 10000
	sh.SetArgs(args[0], rest)
	return sh.Eval(text)
}

// Seed writes the terminal's starting files: the repository's packages under
// GOPATH (so go run and the editor read them), the runnable examples, and a
// first program, script and README in the home directory.
func Seed() error {
	err := fs.WalkDir(algorithm.Source, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := fs.ReadFile(algorithm.Source, p)
		if err != nil {
			return err
		}
		return write(filepath.Join(Library, filepath.FromSlash(p)), string(b))
	})
	if err != nil {
		return err
	}
	for name, body := range starters {
		if err := write(filepath.Join(Home, name), body); err != nil {
			return err
		}
	}
	return nil
}

func write(path, body string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(body), 0o644)
}

// starters are the home directory's first files.
var starters = map[string]string{
	"README.md": strings.TrimLeft(`
# The alGOrithm editor

Every algorithm and data structure is under ~/go/src/github.com/braswelljr/alGOrithm/pkg.
Open one, change it, and run a program that uses it: the edit is what runs.

    go run main.go     run the program here (Go runs on the Go Playground)
    bash hello.sh      run a shell script
    make               run the Makefile's first target (make hello, make run)
    lsx, catx, grepx   every gnux tool, and the GNU names (ls, cat, grep, ...)
    code FILE          open a file in the editor
    help               the shell's built-ins
`, "\n"),
	"main.go": `package main

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/sorting"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

func main() {
	values := []int{5, 3, 9, 1, 7}
	var log trace.Log
	sorted := sorting.TraceQuick(values, &log)
	fmt.Println(sorted)
	fmt.Println(log.Count(trace.Compare), "comparisons,", log.Count(trace.Swap), "swaps")
}
`,
	"hello.sh": `#!/bin/sh
# A shell script, run with: bash hello.sh
echox "Hello from the terminal"
for n in 1 2 3; do
  echox "line $n"
done
lsx ~/go/src/github.com/braswelljr/alGOrithm/pkg/algorithms | headx -5
`,
	"Makefile": "# Run with: make (the first target), make hello, make run\n" +
		".PHONY: all hello run\n\n" +
		"all: hello\n\n" +
		"hello:\n\tbash hello.sh\n\n" +
		"run:\n\tgo run main.go\n",
}
